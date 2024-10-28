package wes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ginWeb/config"
	"ginWeb/service/dataType"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"sync"
	"time"
)

// ws处理超时时间
var handleTimeout = time.Duration(config.Conf.Server.WsTaskTimeout) * time.Second

// ws连接最大时长
var connectionLifeTime = time.Duration(config.Conf.Server.WsLifeTime) * time.Second

// ws连接心跳检测周期
var heartbeat = time.Duration(config.Conf.Server.WsHeartbeat) * time.Second

// Pubs ws已注册订阅事件
var Pubs = make(map[string]Pub)

// PubsLock 已注册订阅事件读写锁
var PubsLock = sync.RWMutex{}

// ws处理逻辑类型
type handleFunc func(ctx *WContext)

// ws已注册处理函数
type handler []handleFunc

// 已注册的ws处理逻辑
var tasks = make(map[string]*handler)

// 接收的ws报文
type payload struct {
	Id        string        `json:"id"`
	Method    string        `json:"method"`
	Params    []interface{} `json:"params"`
	Signature string        `json:"signature"`
}

// ws返回类型
type resp struct {
	Id         string      `json:"id"`
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}

// Pub 事件订阅接口类
type Pub interface {
	// Subscribe 订阅该事件
	Subscribe(connection *Connection)
	// UnSubscribe 取消订阅
	UnSubscribe(connection *Connection)
	// Publish 向收听者发送消息
	Publish([]byte, *Connection)
	// Shutdown 关闭事件
	Shutdown()
}

func (r *resp) String() string {
	return fmt.Sprintf("id:%s statusCode:%d data:%v", r.Id, r.StatusCode, r.Data)
}

// 连接中的用户信息
type userInfo struct {
	userId         int64
	permission     []string
	lifetimeHolder context.Context
	lock           sync.RWMutex
	cancel         context.CancelFunc
}

func newUserInfo() *userInfo {
	return &userInfo{
		userId:         0,
		permission:     make([]string, 0),
		lifetimeHolder: nil,
		cancel:         nil,
		lock:           sync.RWMutex{},
	}
}

// 修改为登录状态
func (u *userInfo) on(id int64, permission []string) {
	u.lock.Lock()
	defer u.lock.Unlock()
	u.permission = permission
	u.userId = id
	ctx, f := context.WithTimeout(context.Background(), time.Duration(config.Conf.Server.WsLoginLifetime)*time.Second)
	u.lifetimeHolder = ctx
	u.cancel = f

	// 登录超时
	go func() {
		select {
		case <-u.lifetimeHolder.Done():
			loguru.SimpleLog(loguru.Debug, "WS", "login expire")
			u.off()
		}
	}()
}

// 修改为未登录状态
func (u *userInfo) off() {
	if !u.status() {
		return
	}

	u.lock.Lock()
	defer u.lock.Unlock()
	u.cancel()
	u.userId = 0
	u.permission = make([]string, 0)
	u.lifetimeHolder = nil
	u.cancel = nil
}

func (u *userInfo) status() bool {
	u.lock.RLock()
	defer u.lock.RUnlock()
	return u.userId != 0
}

// ws格式化日志
func handleLog(code int, ip string, method string, data string, cost time.Duration) {
	info := fmt.Sprintf("%5d | %8s | %20s | %10s | %v", code, cost.String(), ip, method, data)
	if code == 0 {
		loguru.SimpleLog(loguru.Info, "WS", info)
	} else {
		loguru.SimpleLog(loguru.Error, "WS", info)
	}

}

// WContext ws处理上下文
type WContext struct {
	Conn      *Connection
	Request   payload
	attribute map[string]interface{}

	statusCode int
	response   interface{}
	isAbort    bool // 是否已退出
	withResult bool // 是否已设置结果
}

// Result 设置返回结果，终止后续处理逻辑
func (w *WContext) Result(code int, data interface{}) {
	w.statusCode = code
	w.response = data
	w.withResult = true
	w.isAbort = true
}

// 返回数据
func (w *WContext) returnData() {
	if !w.withResult {
		return
	}
	response := resp{
		Id:         w.Request.Id,
		StatusCode: w.statusCode,
		Data:       w.response,
	}
	data, flag := json.Marshal(response)
	if flag != nil {
		handleLog(dataType.WrongData, w.Conn.conn.RemoteAddr().String(), w.Request.Method, "wrong return data", 0)
	}
	_ = w.Conn.Send(data)
}

func (w *WContext) Abort() {
	w.isAbort = true
}

// Get 获取上下文变量
func (w *WContext) Get(k string) (interface{}, bool) {
	attr, exist := w.attribute[k]
	return attr, exist
}

// Set 设置上下文变量
func (w *WContext) Set(k string, v interface{}) {
	w.attribute[k] = v
}

func (w *WContext) handle() {
	defer func() {
		if err := recover(); err != nil {
			loguru.SimpleLog(loguru.Error, "WS", fmt.Sprintf("panic from ws handle: %s", err))
		}
	}()
	if funcs, ok := tasks[w.Request.Method]; ok {
		c := make(chan struct{}, 1)

		// 任务处理计时器
		timeoutCtx, cf := context.WithTimeout(context.Background(), time.Second*time.Duration(config.Conf.Server.WsTaskTimeout))
		defer cf()
		// 执行处理逻辑
		go func() {
			start := time.Now()
			for _, f := range *funcs {
				if w.isAbort {
					break
				}
				f(w)
			}
			cost := time.Since(start)
			handleLog(w.statusCode, w.Conn.conn.RemoteAddr().String(), w.Request.Method, "success", cost)
			// 处理完成信号
			defer func() {
				c <- struct{}{}
			}()
		}()

		select {
		// 处理逻辑超时
		case <-timeoutCtx.Done():
			w.Result(dataType.Timeout, "timeout")
			handleLog(dataType.Timeout, w.Conn.conn.RemoteAddr().String(), w.Request.Method, "timeout", handleTimeout)
		// 获取到完成信号
		case <-c:
			break
		}

	} else {
		w.Result(dataType.NotFound, "not found")
		handleLog(1, w.Conn.conn.RemoteAddr().String(), w.Request.Method, "notfound", 0)
	}
	w.returnData()
}

// NewWContext 创建ws上下文
func NewWContext(conn *Connection, data payload) *WContext {
	return &WContext{Conn: conn, Request: data, attribute: make(map[string]interface{})}
}

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
	conn *websocket.Conn
	// 生命周期上下文
	lifetimeCtx  *context.Context
	cancel       context.CancelFunc
	lock         *sync.RWMutex // 对象读写锁
	wsWriterLock *sync.Mutex   // websocket消息写入锁
	msgChan      chan []byte   // 信息信道
	heartChan    chan int64    // 心跳监测信道
	closed       bool
	// 登录信息
	userInfo *userInfo
	// 连接创建时间
	connectTime time.Time
	// 连接频道
	Channel Pub
}

// NewConnection 新建连接对象
func NewConnection(conn *websocket.Conn) *Connection {
	// 创建生命周期管理上下文
	ctx, cancel := context.WithTimeout(context.Background(), connectionLifeTime)
	c := &Connection{
		conn:         conn,
		lifetimeCtx:  &ctx,
		cancel:       cancel,
		msgChan:      make(chan []byte),
		lock:         &sync.RWMutex{},
		wsWriterLock: &sync.Mutex{},
		userInfo:     newUserInfo(),
		connectTime:  time.Now(),
		Channel:      nil,
	}
	loguru.SimpleLog(loguru.Info, "WS", fmt.Sprintf("connected from %s", conn.RemoteAddr()))
	return c
}

func (c *Connection) Send(data []byte) error {
	if c.IsClose() {
		return errors.New("connection is closed")
	}
	c.wsWriterLock.Lock()
	defer c.wsWriterLock.Unlock()
	err := c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Login 连接登录
func (c *Connection) Login(id int64, permissions []string) {
	c.userInfo.on(id, permissions)
}

func (c *Connection) Logout() {
	c.userInfo.off()
}

// IsLogin 判断是否是登录状态
func (c *Connection) IsLogin() bool {
	return c.userInfo.status()
}

func (c *Connection) IsClose() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.closed
}

// 开始接收数据
func (c *Connection) listen() {
	for {
		if c.IsClose() {
			break
		}
		// 读取失败，被动关闭连接
		type_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.Disconnect()
			break
		}
		switch type_ {
		case websocket.TextMessage:
			c.msgChan <- message
		case websocket.BinaryMessage:
			c.msgChan <- message
		case websocket.CloseMessage:
			c.Disconnect()
		case websocket.PingMessage:
			_ = c.conn.WriteMessage(websocket.PongMessage, []byte("pong"))
			loguru.SimpleLog(loguru.Info, "WS", "pong to "+c.conn.RemoteAddr().String())
		default:
			continue
		}
	}
	loguru.SimpleLog(loguru.Debug, "WS", fmt.Sprintf("close listen from %s", c.conn.RemoteAddr().String()))
}

// 开始处理请求
func (c *Connection) handle() {
	for {
		select {
		// 生命周期结束
		case <-(*c.lifetimeCtx).Done():
			c.Disconnect()
			loguru.SimpleLog(loguru.Debug, "WS", fmt.Sprintf("close handle from %s by lifetime over", c.conn.RemoteAddr().String()))
			return
		case msg := <-c.msgChan:
			var data payload
			err := json.Unmarshal(msg, &data)
			// 错误请求格式
			if err != nil {
				msg := fmt.Sprintf("wrong message: %s", string(msg))
				res, _ := json.Marshal(resp{
					Id:         "",
					StatusCode: dataType.WrongBody,
					Data:       msg,
				})
				_ = c.Send(res)
				handleLog(dataType.WrongBody, c.conn.RemoteAddr().String(), "-", msg, 0)
			} else {
				// 单个请求的处理
				ctx := NewWContext(c, data)
				go ctx.handle()
			}
		}
	}
}

// 处理心跳检测返回信息，10秒超时关闭连接
func (c *Connection) waitHeartbeat(tick int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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
		case <-ctx.Done():
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
				_ = c.conn.WriteMessage(websocket.TextMessage, []byte("ping"))
				go c.waitHeartbeat(t.Unix())
			case <-(*c.lifetimeCtx).Done():
				return
			}
		}
	}()
}

// Disconnect 关闭连接
func (c *Connection) Disconnect() {
	if c.IsClose() {
		return
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.Channel != nil {
		c.Channel.UnSubscribe(c)
	}
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			loguru.SimpleLog(loguru.Error, "WS", "connect close err: "+err.Error())
		}
	}
	// 主动取消生命周期上下文
	c.userInfo.off()
	c.cancel()
	c.closed = true
	loguru.SimpleLog(loguru.Info, "WS", fmt.Sprintf("disconnect from %s, lifetime %s",
		c.conn.RemoteAddr().String(), time.Now().Sub(c.connectTime).String()))
}

// UpgradeConn ws路由升级函数
func UpgradeConn(c *gin.Context) {
	conn, err := upper.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: "upgrade failed",
		})
		return
	}
	// 设置ping响应
	conn.SetPingHandler(func(appData string) error {
		loguru.SimpleLog(loguru.Debug, "WS", fmt.Sprintf("receive ping data '%s' from: %s", appData, conn.RemoteAddr().String()))
		return conn.WriteMessage(websocket.TextMessage, []byte("pong"))
	})
	connect := NewConnection(conn)
	wsChan := c.Query("channel")
	if wsChan != "" {
		pub, flag := Pubs[wsChan]
		if flag {
			connect.Channel = pub
			pub.Subscribe(connect)
		} else {
			c.AbortWithStatusJSON(200, dataType.JsonWrong{
				Code: dataType.WrongData, Message: "channel not found",
			})
			return
		}
	}
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

// RegisterHandler 注册ws处理函数，key已存在则跳过
func RegisterHandler(key string, f ...handleFunc) {
	if _, flag := tasks[key]; flag {
		loguru.Logger.Fatalf("duplicate register ws handler: %s", key)
		return
	}
	var h handler = f
	tasks[key] = &h

}
