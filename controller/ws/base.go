package ws

import (
	"encoding/json"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"time"
)

type Base struct {
}

func (b Base) Ping(w *wes.WContext) {
	if len(w.Request.Params) < 1 {
		w.Result(dataType.WrongBody, "wrong ping body")
		return
	}
	var dt int
	err := json.Unmarshal(w.Request.Params[0], &dt)
	if err != nil {
		w.Result(dataType.WrongData, "wrong ping data: "+err.Error())
		return
	}
	w.Result(dataType.Success, dt)
}

func (b Base) ServerTime(w *wes.WContext) {
	w.Result(dataType.Success, time.Now().UnixMilli())
}

// ConnectUuid 获取当前连接UUID
func (b Base) ConnectUuid(w *wes.WContext) {
	w.Result(dataType.Success, w.Conn.Uuid)
}

func (b Base) RegisterWSRoute(r string, g *wes.Group) {
	group := g.Group(r)
	group.Register("ping", b.Ping)
	group.Register("time", b.ServerTime)
	group.Register("connectUuid", b.ConnectUuid)
}
