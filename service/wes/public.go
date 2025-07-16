package wes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ginWeb/config"
	"ginWeb/model"
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"sync"
	"time"
)

// ws处理超时时间
var handleTimeout = time.Duration(config.Conf.Server.Websocket.WsTaskTimeout) * time.Second

// ws连接最大时长
var connectionLifeTime = time.Duration(config.Conf.Server.Websocket.WsLifeTime) * time.Second

// ws连接心跳检测周期
var heartbeat = time.Duration(config.Conf.Server.Websocket.WsHeartbeat) * time.Second

// 登录超时时间
var loginExpire = time.Duration(config.Conf.Server.Websocket.WsLoginLifetime) * time.Second

// ws处理逻辑类型
type handleFunc func(ctx *WContext)

// ws处理逻辑链
type handler []handleFunc

// 已注册的ws处理逻辑
var tasks = make(map[string]*handler)

// PrintTasks 控制台打印所有ws路由
func PrintTasks() {
	if !config.Conf.Server.Debug {
		return
	}
	for k, v := range tasks {
		fmt.Printf("[WS-debug] %-10s (%v handlers)\n", k, len(*v))
	}
}

// 接收的ws报文
type payload struct {
	Id        string            `json:"id"`
	Method    string            `json:"method"`
	Params    []json.RawMessage `json:"params"`
	Signature string            `json:"signature"`
}

// Resp ws返回类型
type Resp struct {
	Id         string      `json:"id"`
	Method     string      `json:"method"` // 响应类型，reply是被动回复
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}

func (r *Resp) String() string {
	return fmt.Sprintf("id:%s statusCode:%d data:%v", r.Id, r.StatusCode, r.Data)
}

// ws格式化日志
func handleLog(code int, ip string, method string, data string, cost time.Duration) {
	info := fmt.Sprintf("%5d | %8s | %20s | %10s | %v", code, cost.String(), ip, method, data)
	if code == dataType.Success {
		loguru.SimpleLog(loguru.Info, "WS", info)
	} else {
		loguru.SimpleLog(loguru.Error, "WS", info)
	}

}

// WContext ws处理上下文
type WContext struct {
	Conn      *Connection
	Request   *payload
	attribute map[string]interface{}

	statusCode int
	response   interface{}
	isAbort    bool       // 是否已退出
	withResult bool       // 是否已设置结果
	returnOnce *sync.Once // 返回结果的单次锁
}

// Result 设置返回结果，终止后续处理逻辑
func (w *WContext) Result(code int, data interface{}) {
	w.statusCode = code
	w.response = data
	w.withResult = true
	w.isAbort = true
}

// 返回数据，只会执行一次
func (w *WContext) returnData(v []byte) {
	w.returnOnce.Do(func() {
		if v != nil {
			_ = w.Conn.Send(v)
			return
		}
		if !w.withResult {
			return
		}
		response := Resp{
			Id:         w.Request.Id,
			Method:     "reply",
			StatusCode: w.statusCode,
			Data:       w.response,
		}
		data, flag := json.Marshal(response)
		if flag != nil {
			handleLog(dataType.WrongData, w.Conn.conn.RemoteAddr().String(), w.Request.Method, "wrong return data", 0)
		}
		_ = w.Conn.Send(data)
	})
}

// Abort 中断处理
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
			loguru.SimpleLog(loguru.Error, "WS", fmt.Sprintf("panic from ws handle: %v", err))
			w.Result(dataType.WsResolveFailed, "resolve failed")
		}
	}()
	if functions, ok := tasks[w.Request.Method]; ok {

		// 任务处理计时器
		doneCtx, done := context.WithCancel(context.Background())
		start := time.Now()
		// 同步执行处理逻辑
		go func() {
			for _, f := range *functions {
				if w.isAbort {
					break
				}
				f(w)
			}
			done()
		}()

		select {
		// 处理完成
		case <-doneCtx.Done():
			cost := time.Since(start)
			logInfo := "success"
			if w.statusCode != dataType.Success {
				logInfo = w.response.(string)
			}
			handleLog(w.statusCode, w.Conn.conn.RemoteAddr().String(), w.Request.Method, logInfo, cost)
			w.returnData(nil)
		// 处理超时
		case <-time.After(handleTimeout):
			handleLog(dataType.Timeout, w.Conn.conn.RemoteAddr().String(), w.Request.Method, "timeout", handleTimeout)
			r := &Resp{
				Id:         w.Request.Id,
				Method:     "reply",
				StatusCode: dataType.Timeout,
				Data:       "timeout",
			}
			v, _ := json.Marshal(r)
			w.returnData(v)
		}

	} else {
		w.Result(dataType.NotFound, "not found")
		handleLog(1, w.Conn.conn.RemoteAddr().String(), w.Request.Method, "notfound", 0)
		w.returnData(nil)
	}
}

// NewWContext 创建ws上下文
func NewWContext(conn *Connection, data *payload) *WContext {
	return &WContext{Conn: conn, Request: data, attribute: make(map[string]interface{}), returnOnce: &sync.Once{}}
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
	Uuid string
	conn *websocket.Conn
	// 生命周期上下文
	lifetimeCtx    *context.Context
	cancel         context.CancelFunc
	lock           *sync.RWMutex // 对象读写锁
	msgChan        chan *payload // 信息信道
	heartChan      chan int64    // 心跳监测信道
	closed         bool
	disconnectOnce *sync.Once // 断开连接单次执行锁

	address    net.Addr
	MacAddress string
	// 登录信息
	UserId         int64
	UserName       string
	UserPermission []string
	// 连接创建时间
	connectTime time.Time
	// 断开连接时的任务
	doneHooks map[string]func()
}

// NewConnection 新建连接对象
func NewConnection(conn *websocket.Conn, user *systemMode.User, macAddr ...string) *Connection {
	// 创建生命周期管理上下文
	ctx, cancel := context.WithTimeout(context.Background(), connectionLifeTime)
	perms, _ := model.GetPermissionById(user.Id)
	c := &Connection{
		conn:           conn,
		Uuid:           uuid.New().String(),
		lifetimeCtx:    &ctx,
		cancel:         cancel,
		msgChan:        make(chan *payload, config.Conf.Server.Websocket.WsMaxWaiting),
		heartChan:      make(chan int64, config.Conf.Server.Websocket.WsMaxWaiting),
		lock:           &sync.RWMutex{},
		connectTime:    time.Now(),
		address:        conn.RemoteAddr(),
		UserId:         user.Id,
		UserName:       user.Username,
		UserPermission: perms,
		disconnectOnce: &sync.Once{},
		doneHooks:      make(map[string]func()),
	}
	if len(macAddr) > 0 {
		c.MacAddress = macAddr[0]
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
	if c.closed {
		return errors.New("connection is closed")
	}
	err := c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.address
}

func (c *Connection) IsClose() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.closed
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
			if string(message) == "pong" {
				go c.checkInMessage(true, message)
				continue
			}
			go c.checkInMessage(false, message)
		case websocket.BinaryMessage:
			loguru.SimpleLog(loguru.Debug, "WS", "ignore binary message from: "+c.address.String())
			continue
		case websocket.CloseMessage:
			c.Disconnect()
		case websocket.PingMessage:
			_ = c.Send([]byte("pong"))
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
}

func (c *Connection) DeleteDoneHook(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.doneHooks, key)
}

// Disconnect 关闭连接
func (c *Connection) Disconnect() {
	c.disconnectOnce.Do(func() {
		for _, hook := range c.doneHooks {
			hook()
		}
		c.lock.Lock()
		defer c.lock.Unlock()
		if c.closed {
			return
		}

		if c.conn != nil {
			err := c.conn.Close()
			if err != nil {
				loguru.SimpleLog(loguru.Trace, "WS", "connect close err: "+err.Error())
			}
		}
		// 主动取消生命周期上下文
		c.cancel()
		c.closed = true
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
	tokenData, ok := c.Get("token")
	if !ok {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.NoToken, Message: "token not found",
		})
		return
	}
	token := tokenData.(*auth.Token)
	user, err := systemMode.GetUserByID(token.UserId)
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.WrongToken, Message: err.Error(),
		})
		return
	}
	connect := NewConnection(conn, user, c.GetHeader("Mac"))
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

// RegisterHandler 注册ws处理函数，key已存则触发panic
func RegisterHandler(key string, f ...handleFunc) {
	if _, flag := tasks[key]; flag {
		loguru.Logger.Fatalf("duplicate register ws handler: %s", key)
		return
	}
	var h handler = f
	tasks[key] = &h

}

// Group ws处理组
type Group struct {
	node    string
	middles []handleFunc
}

var BasicGroup = &Group{node: "", middles: []handleFunc{}}

// Use 添加中间件
func (g *Group) Use(f ...handleFunc) {
	g.middles = append(g.middles, f...)
}

// Group 创建一个子组
func (g *Group) Group(name string, f ...handleFunc) *Group {
	var key string
	if g.node == "" {
		key = name
	} else {
		key = fmt.Sprintf("%s.%s", g.node, name)
	}
	return &Group{
		node:    key,
		middles: append(g.middles, f...),
	}
}

// Register 在组上创建处理函数
func (g *Group) Register(route string, f ...handleFunc) {
	var key string
	if g.node == "" {
		key = route
	} else {
		key = fmt.Sprintf("%s.%s", g.node, route)
	}
	RegisterHandler(key, append(g.middles, f...)...)
}

// NewGroup 新建组
func NewGroup(name string, f ...handleFunc) *Group {
	return &Group{
		node:    name,
		middles: append([]handleFunc{}, f...),
	}
}
