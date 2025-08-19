package wes

import (
	"context"
	"encoding/json"
	"fmt"
	"ginWeb/config"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"ginWeb/utils/loguru"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ws处理超时时间
var handleTimeout = time.Duration(config.Conf.Server.Websocket.WsTaskTimeout) * time.Second

// ws连接最大时长
var connectionLifeTime = time.Duration(config.Conf.Server.Websocket.WsLifeTime) * time.Second

// ws连接心跳检测周期
var heartbeat = time.Duration(config.Conf.Server.Websocket.WsHeartbeat) * time.Second

var upper = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		loguru.SimpleLog(loguru.Error, "WS", fmt.Sprintf("wes from %s error: %s", r.RemoteAddr, reason.Error()))
	},
}

// Connection ws连接对象
type Connection struct {
	Uuid string
	conn *websocket.Conn
	// 生命周期上下文
	lifetimeCtx    *context.Context
	cancel         context.CancelFunc
	lock           *sync.RWMutex // 对象读写锁
	msgChan        chan *payload // 信息信道
	heartChan      chan int64    // 心跳监测信道
	disconnectOnce *sync.Once    // 断开连接单次执行锁

	IP         string
	MacAddress string
	// 登录信息
	UserId         int64
	UserUuid       string
	UserName       string
	UserPermission []string
	AuthExpireTime time.Time
	// 连接创建时间
	connectTime time.Time
	// 断开连接时的任务
	doneHooks  map[string]func()
	hookChain  []string
	doHookOnce *sync.Once
}

// NewConnection 新建连接对象
func NewConnection(conn *websocket.Conn) *Connection {
	// 创建生命周期管理上下文
	ctx, cancel := context.WithTimeout(context.Background(), connectionLifeTime)
	c := &Connection{
		conn:           conn,
		Uuid:           uuid.New().String(),
		lifetimeCtx:    &ctx,
		cancel:         cancel,
		msgChan:        make(chan *payload, config.Conf.Server.Websocket.WsMaxWaiting),
		heartChan:      make(chan int64, config.Conf.Server.Websocket.WsMaxWaiting),
		lock:           &sync.RWMutex{},
		connectTime:    time.Now(),
		IP:             conn.RemoteAddr().String(),
		UserPermission: make([]string, 0),
		disconnectOnce: &sync.Once{},

		doneHooks:  make(map[string]func()),
		hookChain:  make([]string, 0),
		doHookOnce: &sync.Once{},
	}
	loguru.SimpleLog(loguru.Info, "WS", fmt.Sprintf("connected from %s", conn.RemoteAddr()))
	return c
}

// Done 连接生命周期
func (c *Connection) Done() <-chan struct{} {
	return (*c.lifetimeCtx).Done()
}

// Send 发送消息
func (c *Connection) Send(data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	err := c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}
	return nil
}

// 将消息压入通道，压入失败则返回请求错误
func (c *Connection) checkInMessage(isHeartbeat bool, msg []byte) {

	if isHeartbeat {
		select {
		case c.heartChan <- time.Now().Unix():
			return
		case <-time.After(1 * time.Second):
			var m = Resp{
				Id:         "",
				Method:     "reply",
				StatusCode: dataType.TooManyRequests,
				Data:       "too much heartbeat",
			}
			data, _ := json.Marshal(m)
			_ = c.Send(data)
			return
		}

	}
	var req payload
	err := json.Unmarshal(msg, &req)
	if err != nil {
		msg := fmt.Sprintf("wrong message: %s", string(msg))
		res, _ := json.Marshal(Resp{
			Id:         "",
			Method:     "reply",
			StatusCode: dataType.WrongBody,
			Data:       msg,
		})
		_ = c.Send(res)
		handleLog(dataType.WrongBody, c.conn.RemoteAddr().String(), "-", msg, 0)
		return
	}
	select {
	case c.msgChan <- &req:
		return
	case <-time.After(1 * time.Second):
		var m = Resp{
			Id:         req.Id,
			Method:     "reply",
			StatusCode: dataType.TooManyRequests,
			Data:       "too much request",
		}
		data, _ := json.Marshal(m)
		_ = c.Send(data)
		return
	}
}

// 开始接收数据
func (c *Connection) listen() {
	for {
		// 读取失败，被动关闭连接
		type_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.Disconnect()
			break
		}
		switch type_ {
		case websocket.TextMessage:
			if string(message) == "pong" {
				go c.checkInMessage(true, message)
				continue
			}
			go c.checkInMessage(false, message)
		case websocket.BinaryMessage:
			loguru.SimpleLog(loguru.Debug, "WS", "ignore binary message from: "+c.IP)
			continue
		case websocket.CloseMessage:
			c.Disconnect()
		case websocket.PingMessage:
			_ = c.Send([]byte("pong"))
		default:
			continue
		}
	}
	loguru.SimpleLog(loguru.Debug, "WS", fmt.Sprintf("close listen from %s", c.IP))
}

// 开始处理请求
func (c *Connection) handle() {
	for {
		select {
		// 生命周期结束
		case <-(*c.lifetimeCtx).Done():
			loguru.SimpleLog(loguru.Debug, "WS", fmt.Sprintf("close handle from %s by lifetime over", c.conn.RemoteAddr().String()))
			return
		case msg := <-c.msgChan:
			// 单个请求的处理
			ctx := NewWContext(c, msg)
			go ctx.handle()

		}
	}
}

// 处理心跳检测返回信息，10秒超时关闭连接
func (c *Connection) waitHeartbeat(tick int64) {
	for {
		select {
		case <-(*c.lifetimeCtx).Done():
			c.Disconnect()
			return
		case t := <-c.heartChan:
			// 超过十秒或早于发心跳检测则无效
			if t > tick+10 || t < tick {
				continue
			}
			return
		case <-time.After(10 * time.Second):
			loguru.SimpleLog(loguru.Info, "WS", fmt.Sprintf("heartbeat failed from %s", c.conn.RemoteAddr().String()))
			c.Disconnect()
			return
		}
	}
}

// 启动心跳检测
func (c *Connection) heartbeat() {
	ticker := time.NewTicker(heartbeat)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				_ = c.Send([]byte("ping"))
				go c.waitHeartbeat(t.Unix())
			case <-(*c.lifetimeCtx).Done():
				return
			}
		}
	}()
}

// DoneHook 添加断开连接钩子函数
func (c *Connection) DoneHook(key string, f func()) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.doneHooks[key] = f
	c.hookChain = append(c.hookChain, key)
}

// DeleteDoneHook 删除断连钩子
func (c *Connection) DeleteDoneHook(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.doneHooks, key)
	for idx, hook := range c.hookChain {
		if hook == key {
			if idx == len(c.hookChain)-1 {
				c.hookChain = c.hookChain[:idx]
			} else {
				c.hookChain = append(c.hookChain[:idx], c.hookChain[idx+1:]...)
			}
			break
		}
	}
}

func (c *Connection) doHook() {
	c.doHookOnce.Do(func() {
		for i := len(c.hookChain) - 1; i >= 0; i-- {
			f, ok := c.doneHooks[c.hookChain[i]]
			if !ok {
				continue
			}
			f()
		}
	})
}

// Auth 登录认证，可选传入mac地址
func (c *Connection) Auth(s string, mac ...string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	token, err := auth.CheckToken(s)
	if err != nil {
		return err
	}
	c.UserId = token.UserId
	c.UserName = token.Username
	c.UserPermission = token.Permission
	c.AuthExpireTime = token.Expire
	c.UserUuid = token.UserUUID
	if len(mac) == 1 {
		c.MacAddress = mac[0]
	}
	return nil
}

// Disconnect 关闭连接
func (c *Connection) Disconnect() {
	c.disconnectOnce.Do(func() {
		c.doHook()
		c.lock.Lock()
		defer c.lock.Unlock()

		if c.conn != nil {
			err := c.conn.Close()
			if err != nil {
				loguru.SimpleLog(loguru.Trace, "WS", "connect close err: "+err.Error())
			}
		}
		// 主动取消生命周期上下文
		c.cancel()
		loguru.SimpleLog(loguru.Info, "WS", fmt.Sprintf("disconnect from %s, lifetime %s",
			c.conn.RemoteAddr().String(), time.Now().Sub(c.connectTime).String()))
	})
}

// UpgradeConn ws路由升级函数
func UpgradeConn(c *gin.Context) {
	conn, err := upper.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.Unknown, Message: "upgrade failed",
		})
		return
	}
	// 设置ping响应
	conn.SetPingHandler(func(appData string) error {
		loguru.SimpleLog(loguru.Trace, "WS", fmt.Sprintf("receive ping data '%s' from: %s", appData, conn.RemoteAddr().String()))
		return conn.WriteMessage(websocket.TextMessage, []byte("pong"))
	})
	connect := NewConnection(conn)
	// 设置连接关闭时调用管理对象的Disconnect方法
	conn.SetCloseHandler(func(code int, text string) error {
		connect.Disconnect()
		return nil
	})
	// 开启连接的监听和处理函数
	go connect.listen()
	go connect.handle()
	connect.heartbeat()
}
