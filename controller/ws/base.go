package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/database"
	"time"
)

func ServerTime(w *wes.WContext) {
	w.Result(dataType.Success, time.Now().UnixMilli())
}

func Auth(w *wes.WContext) {
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
	result := database.Rdb.SetNX(context.Background(), fmt.Sprintf("ws::unique::%s", w.Conn.UserUuid), true, 0)
	if r, err := result.Result(); !r || err != nil {
		w.Result(dataType.WrongData, "auth failed")
		return
	}
	w.Conn.DoneHook("breakAuth", func() {
		database.Rdb.Del(context.Background(), fmt.Sprintf("ws::unique::%s", w.Conn.UserUuid))
	})
	w.Result(dataType.Success, "auth success")
}
