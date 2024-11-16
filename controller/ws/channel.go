package ws

import (
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
	"strings"
)

// SubHandle ws订阅事件接口
func SubHandle(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongData, "without params")
		return
	}
	var failedKeys = make([]string, 0)
	for _, name := range w.Request.Params {
		pub, ok := subscribe.GetPub(name)
		if ok {
			_ = pub.Subscribe(w.Conn)
		} else {
			failedKeys = append(failedKeys, name)
		}
	}
	if len(failedKeys) > 0 {
		w.Result(dataType.NotFound, strings.Join(failedKeys, ","))
	}
	w.Result(dataType.Success, "success")
}

// UnsubHandle ws取消事件订阅接口
func UnsubHandle(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongData, "without params")
		return
	}

	for _, name := range w.Request.Params {
		pub, ok := subscribe.GetPub(name)
		if ok {
			_ = pub.UnSubscribe(w.Conn)
		}
	}
	w.Result(dataType.Success, "success")
}

// Broadcast 向频道发送消息
func Broadcast(w *wes.WContext) {

	if len(w.Request.Params) != 2 {
		w.Result(dataType.WrongBody, "invalid param")
		return
	}
	name := w.Request.Params[0]
	msg := w.Request.Params[1]
	pub, ok := subscribe.GetPub(name)
	if !ok {
		w.Result(dataType.NotFound, "not found pub")
		return
	}
	err := pub.Publish(msg, w.Conn)
	if err != nil {
		w.Result(dataType.WrongBody, "broadcast failed:"+err.Error())
		return
	}
	w.Result(dataType.Success, "success")
}
