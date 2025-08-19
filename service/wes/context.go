package wes

import (
	"context"
	"encoding/json"
	"fmt"
	"ginWeb/config"
	"ginWeb/service/dataType"
	"ginWeb/utils/loguru"
	"sync"
	"time"
)

// ws处理逻辑类型
type handleFunc func(ctx *WContext)

// ws处理逻辑链
type handler []handleFunc

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
		loguru.SimpleLog(loguru.Debug, "WS", info)
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
		handleLog(1, w.Conn.conn.RemoteAddr().String(), w.Request.Method, "not found", 0)
		w.returnData(nil)
	}
}

// NewWContext 创建ws上下文
func NewWContext(conn *Connection, data *payload) *WContext {
	return &WContext{Conn: conn, Request: data, attribute: make(map[string]interface{}), returnOnce: &sync.Once{}}
}

// ===================== 连接对象 ==================
