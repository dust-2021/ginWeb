package ws

import (
	"encoding/json"
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
	"github.com/gin-gonic/gin"
	"strings"
)

type ChannelController struct {
}

// SubHandle ws订阅事件接口
func (c ChannelController) SubHandle(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongData, "without params")
		return
	}
	var failedKeys = make([]string, 0)
	for _, name := range w.Request.Params {
		var n string
		err := json.Unmarshal(name, &n)
		if err != nil {
			failedKeys = append(failedKeys, n)
			continue
		}
		pub, ok := subscribe.GetPub(n)
		if ok {
			_ = pub.Subscribe(w.Conn)
		} else {
			failedKeys = append(failedKeys, n)
		}
	}
	if len(failedKeys) > 0 {
		w.Result(dataType.NotFound, strings.Join(failedKeys, ","))
		return
	}
	w.Result(dataType.Success, "success")
}

// UnsubHandle ws取消事件订阅接口
func (c ChannelController) UnsubHandle(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongData, "without params")
		return
	}
	var failedKeys = make([]string, 0)
	for _, name := range w.Request.Params {
		var n string
		err := json.Unmarshal(name, &n)
		if err != nil {
			failedKeys = append(failedKeys, n)
			continue
		}
		pub, ok := subscribe.GetPub(n)
		if ok {
			_ = pub.UnSubscribe(w.Conn)
		} else {
			failedKeys = append(failedKeys, n)
		}
	}
	if len(failedKeys) > 0 {
		w.Result(dataType.NotFound, strings.Join(failedKeys, ","))
		return
	}
	w.Result(dataType.Success, "success")
}

// Broadcast 向频道发送消息
func (c ChannelController) Broadcast(w *wes.WContext) {

	if len(w.Request.Params) != 2 {
		w.Result(dataType.WrongBody, "invalid param")
		return
	}
	var name string
	var msg string
	err := json.Unmarshal(w.Request.Params[0], &name)
	if err != nil {
		w.Result(dataType.WrongBody, "invalid name")
		return
	}
	err = json.Unmarshal(w.Request.Params[1], &msg)
	if err != nil {
		w.Result(dataType.WrongBody, "invalid msg")
	}
	pub, ok := subscribe.GetPub(name)
	if !ok {
		w.Result(dataType.NotFound, "not found pub")
		return
	}
	pub.Message(msg, w.Conn)
	w.Result(dataType.Success, "success")
}

func (c ChannelController) RegisterRoute(r string, g *gin.RouterGroup) {}

func (c ChannelController) RegisterWSRoute(r string, g *wes.Group) {

	group := g.Group(r)

	group.Register("broadcast", middleware.AuthMiddle.WsHandle,
		middleware.NewPermission([]string{"channel.broadcast"}).WsHandle, c.Broadcast)
	group.Register("subscribe", c.SubHandle)
	group.Register("unsubscribe", c.UnsubHandle)
}
