package ws

import (
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RoomController struct {
}

func (r RoomController) CreateRoom(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongBody, "Room Create Failed without title")
		return
	}
	roomName := w.Request.Params[0]
	room, err := subscribe.NewRoom(w.Conn, roomName)
	if err != nil {
		w.Result(dataType.Unknown, err.Error())
		return
	}
	w.Result(dataType.Success, room.UUID())
}

func (r RoomController) GetInRoom(w *wes.WContext) {
	if len(w.Request.Params) != 1 {
		w.Result(dataType.WrongBody, "invalid params")
		return
	}
	roomId := w.Request.Params[0]
	room, ok := subscribe.GetRoom(roomId)
	if !ok {
		w.Result(dataType.NotFound, "not found")
		return
	}
	err := room.Subscribe(w.Conn)
	if err != nil {
		w.Result(dataType.Unknown, "subscribe failed")
		return
	}
	w.Result(dataType.Success, "success")
}

func (r RoomController) GetOutRoom(w *wes.WContext) {
	if len(w.Request.Params) != 1 {
		w.Result(dataType.WrongBody, "invalid params")
		return
	}
	roomId := w.Request.Params[0]
	room, ok := subscribe.GetRoom(roomId)
	if !ok {
		w.Result(dataType.NotFound, "not found")
		return
	}
	err := room.UnSubscribe(w.Conn)
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
	roomId := w.Request.Params[0]
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
	roomId := w.Request.Params[0]
	room, ok := subscribe.GetRoom(roomId)
	if !ok {
		w.Result(dataType.NotFound, "not found")
		return
	}
	if !room.ExistMember(w.Conn) {
		w.Result(dataType.DeniedByPermission, "not in room")
		return
	}
	resp := struct {
		RoomUUID string               `json:"roomUuid"`
		RoomName string               `json:"RoomName"`
		RoomMate []subscribe.MateInfo `json:"RoomMate"`
	}{
		RoomUUID: room.UUID(),
		RoomName: room.Title,
		RoomMate: room.Mates(),
	}
	w.Result(dataType.Success, resp)
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
