package ws

import (
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
func (c ChannelController) UnsubHandle(w *wes.WContext) {
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
func (c ChannelController) Broadcast(w *wes.WContext) {

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

func (c ChannelController) RegisterRoute(r string, g *gin.RouterGroup) {}

func (c ChannelController) RegisterWSRoute(r string, g *wes.Group) {

	group := g.Group(r)

	group.Register("broadcast", middleware.NewLoginStatus().WsHandle,
		middleware.NewPermission([]string{"channel.broadcast"}).WsHandle, c.Broadcast)
	group.Register("subscribe", c.SubHandle)
	group.Register("unsubscribe", c.UnsubHandle)
}
