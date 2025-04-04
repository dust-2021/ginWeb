package ws

import (
	"encoding/json"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
	"ginWeb/utils/tools"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RoomController struct {
}

func (r RoomController) CreateRoom(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongBody, "Room Create Failed without config")
		return
	}
	var conf subscribe.RoomConfig
	err := tools.ShouldBindJson(w.Request.Params[0], &conf)
	if err != nil {
		w.Result(dataType.WrongBody, err.Error())
		return
	}
	room, err := subscribe.NewRoom(w.Conn, &conf)
	if err != nil {
		w.Result(dataType.Unknown, err.Error())
		return
	}
	w.Result(dataType.Success, room.UUID())
}

func (r RoomController) GetInRoom(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongBody, "invalid params")
		return
	}
	var roomId string
	err := json.Unmarshal(w.Request.Params[0], &roomId)
	if err != nil {
		w.Result(dataType.WrongBody, "invalided room id")
		return
	}
	var password = ""
	if len(w.Request.Params) > 1 {
		var p string
		err := json.Unmarshal(w.Request.Params[1], &p)
		if err != nil {
			w.Result(dataType.WrongBody, "invalided password")
			return
		}
		password = p
	}
	room, ok := subscribe.GetRoom(roomId)
	if !ok {
		w.Result(dataType.WrongBody, "room not found")
		return
	}
	if password != room.Config.Password {
		w.Result(dataType.DeniedByPermission, "invalid password")
		return
	}
	if !ok {
		w.Result(dataType.NotFound, "not found")
		return
	}
	err = room.Subscribe(w.Conn)
	if err != nil {
		w.Result(dataType.Unknown, "subscribe failed: "+err.Error())
		return
	}
	w.Result(dataType.Success, "success")
}

func (r RoomController) GetOutRoom(w *wes.WContext) {
	if len(w.Request.Params) != 1 {
		w.Result(dataType.WrongBody, "invalid params")
		return
	}
	var roomId string
	err := json.Unmarshal(w.Request.Params[0], &roomId)
	if err != nil {
		w.Result(dataType.WrongBody, "invalided room id")
		return
	}
	room, ok := subscribe.GetRoom(roomId)
	if !ok {
		w.Result(dataType.NotFound, "not found")
		return
	}
	err = room.UnSubscribe(w.Conn)
	if err != nil {
		w.Result(dataType.Unknown, "subscribe failed")
		return
	}
	w.Result(dataType.Success, "success")
}

func (r RoomController) CloseRoom(w *wes.WContext) {
	if len(w.Request.Params) != 1 {
		w.Result(dataType.WrongBody, "invalid params")
		return
	}
	var roomId string
	err := json.Unmarshal(w.Request.Params[0], &roomId)
	if err != nil {
		w.Result(dataType.WrongBody, "invalided room id")
		return
	}
	room, ok := subscribe.GetRoom(roomId)
	if !ok {
		w.Result(dataType.NotFound, "not found")
		return
	}
	if room.Owner.UserId != w.Conn.UserId {
		w.Result(dataType.DeniedByPermission, "you are not room owner")
		return
	}
	_ = room.Shutdown()
	w.Result(dataType.Success, "success")
}

func (r RoomController) RoomMate(w *wes.WContext) {
	if len(w.Request.Params) != 1 {
		w.Result(dataType.WrongBody, "invalid params")
		return
	}
	var roomId string
	err := json.Unmarshal(w.Request.Params[0], &roomId)
	if err != nil {
		w.Result(dataType.WrongBody, "invalided room id")
		return
	}
	room, ok := subscribe.GetRoom(roomId)
	if !ok {
		w.Result(dataType.NotFound, "not found")
		return
	}
	if !room.ExistMember(w.Conn) {
		w.Result(dataType.DeniedByPermission, "not in room")
		return
	}
	w.Result(dataType.Success, room.Mates())
}

func (r RoomController) ListRoom(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusOK, dataType.JsonRes{
		Code: dataType.Success,
		Data: subscribe.RoomInfo(),
	})
}

func (r RoomController) RegisterRoute(route string, g *gin.RouterGroup) {
	g.Group(route).Handle("GET", "list", r.ListRoom)

}

func (r RoomController) RegisterWSRoute(route string, g *wes.Group) {
	group := g.Group(route)
	group.Register("in", r.GetInRoom)
	group.Register("out", r.GetOutRoom)
	group.Register("close", r.CloseRoom)
	group.Register("roommate", r.RoomMate)
	group.Register("create", r.CreateRoom)
}
