package wes

import (
	"context"
	"encoding/json"
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

// ws处理逻辑类型
type handleFunc func(ctx *WContext) interface{}

// 已注册的ws处理逻辑
var tasks map[string]handleFunc

// 接收的ws报文
type payload struct {
	Id        string   `json:"id"`
	Method    string   `json:"method"`
	Params    []string `json:"params"`
	Signature string   `json:"signature"`
}

// ws返回类型
type resp struct {
	Id         string      `json:"id"`
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"req"`
}

func (r *resp) String() string {
	return fmt.Sprintf("id:%s statusCode:%d data:%v", r.Id, r.StatusCode, r.Data)
}

// WContext ws处理上下文
type WContext struct {
	conn      *websocket.Conn
	req       payload
	attribute map[string]interface{}
	ctx       context.Context
	cancel    context.CancelFunc
}

// 设置返回结果并发送
func (w *WContext) setResult(code int, data interface{}) {
	res := resp{
		Id:         w.req.Id,
		StatusCode: code,
		Data:       data,
	}
	msg, err := json.Marshal(res)
	if err != nil {
		loguru.Logger.Errorf("ws resp cant be serialized: %s", res.String())
		return
	}
	_ = w.conn.WriteMessage(websocket.TextMessage, msg)
	loguru.Logger.Infof(" [WS] | %5d | %10s | %10s | %s", code, w.conn.RemoteAddr().String(), w.req.Method, res.String())
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
	if f, ok := tasks[w.req.Method]; ok {
		c := make(chan interface{}, 1)
		// 执行处理逻辑
		go func() {
			c <- f(w)
		}()
		defer w.cancel()
		for {
			select {
			case <-w.ctx.Done():
				w.setResult(1, "timeout")
				return
			case data := <-c:
				w.setResult(0, data)
				return
			}

		}
	} else {
		w.setResult(404, "not found")
	}

}

// 创建ws上下文
func newWContext(conn *websocket.Conn, data payload, timeout uint32) *WContext {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	return &WContext{conn: conn, req: data, attribute: make(map[string]interface{}), ctx: ctx, cancel: cancel}
}

// 根据报文分配处理函数
func handlePayload(conn *websocket.Conn, message []byte) {
	var data payload
	err := json.Unmarshal(message, &data)
	if err != nil {
		msg := fmt.Sprintf("wrong message: %s", string(message))
		res, _ := json.Marshal(resp{
			Id:         "",
			StatusCode: 1,
			Data:       msg,
		})
		_ = conn.WriteMessage(websocket.TextMessage, res)
		loguru.Logger.Infof(" [WS] | %5d | %10s | %s", 1, conn.RemoteAddr().String(), msg)
		return
	}
	ctx := newWContext(conn, data, 10)
	ctx.handle()
}

var upper = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		loguru.Logger.Errorf("wes from %s error: %s", r.RemoteAddr, reason.Error())
	},
}

func UpgradeConn(c *gin.Context) {
	conn, err := upper.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: "upgrade failed",
		})
		return
	}
	for {

		t, message, err := conn.ReadMessage()
		if t == websocket.CloseMessage {
			loguru.Logger.Infof("close from %s", conn.RemoteAddr())
			break
		}
		if err != nil {
			resp := map[string]interface{}{}
			data, _ := json.Marshal(resp)
			_ = conn.WriteMessage(t, data)
			err := conn.Close()
			if err != nil {
				loguru.Logger.Errorf("close failed %s", err.Error())
			}
			break
		}
		go handlePayload(conn, message)

	}
}

func init() {
}
