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

// ws处理超时时间
var handleTimeout = 2 * time.Second

// ws处理逻辑类型
type handleFunc func(ctx *WContext) (interface{}, error)

// 已注册的ws处理逻辑
var tasks map[string]handleFunc

// 接收的ws报文
type payload struct {
	Id        string   `json:"id"`
	Method    string   `json:"method"`
	Params    []string `json:"params"`
	Signature string   `json:"signature"`
}

// ws格式化日志
func handleLog(code int, ip string, method string, data string, cost time.Duration) {
	if code == 0 {
		loguru.Logger.Infof(" [WS] | %5d | %8s | %20s | %10s | %v", code, cost.String(), ip, method, data)
	} else {
		loguru.Logger.Errorf(" [WS] | %5d | %8s | %20s | %10s | %v", code, cost.String(), ip, method, data)
	}

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
	ctx       *context.Context
	cancel    context.CancelFunc
	signal    bool
}

// 设置返回结果并发送
func (w *WContext) setResult(code int, data interface{}) error {
	res := resp{
		Id:         w.req.Id,
		StatusCode: code,
		Data:       data,
	}
	msg, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return w.conn.WriteMessage(websocket.TextMessage, msg)
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
		c := make(chan struct{}, 1)
		start := time.Now()
		// 执行处理逻辑
		go func() {
			result, err := f(w)

			var sendErr error
			if err != nil {
				sendErr = w.setResult(dataType.WsResolveFailed, err.Error())
			} else {
				sendErr = w.setResult(dataType.Success, result)
			}

			// 设置响应错误失败，日志记录异常
			if sendErr != nil {
				loguru.Logger.Errorf("ws send error: %s", sendErr.Error())
			}
			c <- struct{}{}
		}()

		select {
		case <-(*w.ctx).Done():
			defer w.cancel()
			_ = w.setResult(dataType.Timeout, "timeout")
			handleLog(dataType.Timeout, w.conn.RemoteAddr().String(), w.req.Method, "timeout", handleTimeout)
			return
		case <-c:
			cost := time.Since(start)
			handleLog(0, w.conn.RemoteAddr().String(), w.req.Method, "success", cost)
			return
		}

	} else {
		err := w.setResult(dataType.NotFound, "not found")
		var msg string
		if err != nil {
			msg = err.Error()
		} else {
			msg = "not found"
		}
		handleLog(1, w.conn.RemoteAddr().String(), w.req.Method, msg, 0)
	}

}

// NewWContext 创建ws上下文
func NewWContext(conn *websocket.Conn, data payload, timeout uint32) *WContext {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	return &WContext{conn: conn, req: data, attribute: make(map[string]interface{}), ctx: &ctx, cancel: cancel}
}

// 根据报文分配处理函数
func handlePayload(conn *websocket.Conn, message []byte) {
	var data payload
	err := json.Unmarshal(message, &data)

	// 获取到异常格式报文
	if err != nil {
		msg := fmt.Sprintf("wrong message: %s", string(message))
		res, _ := json.Marshal(resp{
			Id:         "",
			StatusCode: dataType.WrongBody,
			Data:       msg,
		})
		_ = conn.WriteMessage(websocket.TextMessage, res)
		handleLog(dataType.WrongBody, conn.RemoteAddr().String(), "-", msg, 0)
		return
	}
	ctx := NewWContext(conn, data, 10)
	ctx.handle()
}

var upper = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		loguru.Logger.Errorf("wes from %s error: %s", r.RemoteAddr, reason.Error())
	},
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
	for {
		t, message, err := conn.ReadMessage()
		//if t == websocket.CloseMessage {
		//	loguru.Logger.Infof("close from %s", conn.RemoteAddr())
		//	break
		//}
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

func RegisterHandler(key string, f handleFunc) {
	if _, flag := tasks[key]; flag {
		return
	}
	tasks[key] = f
}

func init() {
	tasks = make(map[string]handleFunc)
	RegisterHandler("hello", func(ctx *WContext) (interface{}, error) {
		return "hello", nil
	})
}
