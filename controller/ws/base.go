package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wireguard"
	"ginWeb/utils/database"
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

// Auth ws认证
// params: [token: string, mac?: string]
func (b Base) Auth(w *wes.WContext) {
	if len(w.Request.Params) < 1 {
		w.Result(dataType.WrongBody, "wrong body")
		return
	}
	var token string
	err := json.Unmarshal(w.Request.Params[0], &token)
	if err != nil {
		w.Result(dataType.WrongData, "wrong token: "+err.Error())
		return
	}
	var mac = ""
	if len(w.Request.Params) == 2 {
		err := json.Unmarshal(w.Request.Params[1], &mac)
		if err != nil {
			w.Result(dataType.WrongData, "wrong mac: "+err.Error())
			return
		}
	}
	err = w.Conn.Auth(token, mac)
	if err != nil {
		w.Result(dataType.WrongData, "auth failed: "+err.Error())
		return
	}
	uuid := w.Conn.AuthInfo().UserUUID
	result := database.Rdb.SetNX(context.Background(), fmt.Sprintf("ws::unique::%s", uuid), true, 0)
	if r, err := result.Result(); !r || err != nil {
		w.Result(dataType.WsDuplicateAuth, "auth failed")
		return
	}
	w.Conn.DoneHook("breakAuth", func() {
		database.Rdb.Del(context.Background(), fmt.Sprintf("ws::unique::%s", uuid))
	})
	w.Result(dataType.Success, "auth success")
}

// ConnectUuid 获取当前连接UUID
func (b Base) ConnectUuid(w *wes.WContext) {
	w.Result(dataType.Success, w.Conn.Uuid)
}

func (b Base) ServerPublicKey(w *wes.WContext) {
	w.Result(dataType.Success, wireguard.WireguardManager.GetPublicKey())
}

func (b Base) RegisterWSRoute(r string, g *wes.Group) {
	group := g.Group(r)
	group.Register("ping", b.Ping)
	group.Register("time", b.ServerTime)
	group.Register("auth", b.Auth)
	group.Register("connectUuid", b.ConnectUuid)
	group.Register("serverPublicKey", b.ServerPublicKey)
}
