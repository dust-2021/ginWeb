package ws

import (
	"encoding/json"
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
	"ginWeb/utils/tools"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type RoomController struct {
}

// CreateRoom 创建房间
// params: [config: subscribe.RoomConfig]
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

// GetInRoom 进入房间
// params: [rooId: string, password?: string]
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
	room, ok := subscribe.Roomer.Get(roomId)
	if !ok {
		w.Result(dataType.WrongBody, "room not found")
		return
	}
	var password = ""
	if len(w.Request.Params) > 1 {
		var p string
		err := json.Unmarshal(w.Request.Params[1], &p)
		if err != nil || len(p) > 16 {
			w.Result(dataType.WrongBody, "invalided password")
			return
		}
		password = p
	}

	if password != *room.Config.Password {
		w.Result(dataType.DeniedByPermission, "invalid password")
		return
	}
	err = room.Subscribe(w.Conn)
	if err != nil {
		w.Result(dataType.Unknown, "subscribe failed: "+err.Error())
		return
	}
	w.Result(dataType.Success, "success")
}

// GetOutRoom 退出房间
// params: [roomId: string]
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
	room, ok := subscribe.Roomer.Get(roomId)
	if !ok {
		w.Result(dataType.NotFound, "room not found")
		return
	}
	err = room.UnSubscribe(w.Conn)
	if err != nil {
		w.Result(dataType.Unknown, "subscribe failed")
		return
	}
	w.Result(dataType.Success, "success")
}

// CloseRoom 关闭房间
// params: [roomId: string]
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
	room, ok := subscribe.Roomer.Get(roomId)
	if !ok {
		w.Result(dataType.NotFound, "room not found")
		return
	}
	if room.Owner.UserId != w.Conn.UserId {
		w.Result(dataType.DeniedByPermission, "you are not room owner")
		return
	}
	_ = room.Shutdown()
	w.Result(dataType.Success, "success")
}

// ForbiddenRoom 房间禁止进入
// params: [rooId: string, stat: bool]
func (r RoomController) ForbiddenRoom(w *wes.WContext) {
	if len(w.Request.Params) != 2 {
		w.Result(dataType.WrongBody, "invalid params")
		return
	}
	var roomId string
	var forbidden bool
	err := json.Unmarshal(w.Request.Params[0], &roomId)
	err2 := json.Unmarshal(w.Request.Params[1], &forbidden)
	if err != nil || err2 != nil {
		w.Result(dataType.WrongBody, "invalided room id")
		return
	}
	room, ok := subscribe.Roomer.Get(roomId)
	if !ok {
		w.Result(dataType.NotFound, "room not found")
	}
	if room.Owner.UserId != w.Conn.UserId {
		w.Result(dataType.DeniedByPermission, "you are not room owner")
		return
	}
	room.Forbidden(forbidden)
	w.Result(dataType.Success, "success")
}

// RoomMate 获取房间成员
// params: [roomId: string]
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
	room, ok := subscribe.Roomer.Get(roomId)
	if !ok {
		w.Result(dataType.NotFound, "room not found")
		return
	}
	if !room.ExistMember(w.Conn) {
		w.Result(dataType.DeniedByPermission, "not in room")
		return
	}
	w.Result(dataType.Success, room.Mates())
}

// RoomMessage 发送消息
// params: [roomId: string, message: string]
func (r RoomController) RoomMessage(w *wes.WContext) {
	if len(w.Request.Params) != 2 {
		w.Result(dataType.WrongBody, "invalid params")
		return
	}
	var roomId string
	err := json.Unmarshal(w.Request.Params[0], &roomId)
	if err != nil {
		w.Result(dataType.WrongBody, "invalided room id")
		return
	}
	var message string
	err = json.Unmarshal(w.Request.Params[1], &message)
	if err != nil {
		w.Result(dataType.WrongBody, "invalided room message")
		return
	}
	room, ok := subscribe.Roomer.Get(roomId)
	if !ok {
		w.Result(dataType.NotFound, "room not found")
		return
	}
	if !room.ExistMember(w.Conn) {
		w.Result(dataType.DeniedByPermission, "not in room")
		return
	}
	room.Message(message, w.Conn)
	w.Result(dataType.Success, "success")
}

// Nat 向目标IP发起nat请求
// params: [rooId: string, key: string, targets: []string]
func (r RoomController) Nat(w *wes.WContext) {
	if len(w.Request.Params) != 3 {
		w.Result(dataType.WrongBody, "invalid params")
		return
	}
	var roomId string
	var key string
	var targets []string
	err := json.Unmarshal(w.Request.Params[0], &roomId)
	if err != nil {
		w.Result(dataType.WrongBody, "invalided room id")
		return
	}
	err = json.Unmarshal(w.Request.Params[1], &key)
	if err != nil {
		w.Result(dataType.WrongBody, "invalided key")
		return
	}
	err = json.Unmarshal(w.Request.Params[2], &targets)
	if err != nil {
		w.Result(dataType.WrongBody, "invalided targets")
		return
	}
	room, ok := subscribe.Roomer.Get(roomId)
	if !ok {
		w.Result(dataType.NotFound, "room not found")
		return
	}
	for _, target := range targets {
		room.Nat(target, key)
	}
}

// ListRoom 所有房间信息接口
func (r RoomController) ListRoom(c *gin.Context) {
	type respInfo struct {
		Total int                  `json:"total"`
		Rooms []subscribe.RoomInfo `json:"rooms"`
	}

	pageNum, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, dataType.JsonRes{
			Code: dataType.WrongBody, Data: "invalided page",
		})
		return
	}

	pageSize, err := strconv.Atoi(c.Query("size"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusOK, dataType.JsonRes{
			Code: dataType.WrongBody, Data: "invalided page",
		})
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, dataType.JsonRes{
		Code: dataType.Success,
		Data: respInfo{
			Total: subscribe.Roomer.Size(),
			Rooms: subscribe.Roomer.List(pageNum, pageSize),
		},
	})
}

func (r RoomController) RegisterRoute(route string, g *gin.RouterGroup) {
	g.Group(route).Handle("GET", "list", r.ListRoom)

}

func (r RoomController) RegisterWSRoute(route string, g *wes.Group) {
	group := g.Group(route)
	group.Use(middleware.NewLoginStatus().WsHandle)
	group.Register("in", r.GetInRoom)
	group.Register("out", r.GetOutRoom)
	group.Register("close", r.CloseRoom)
	group.Register("forbidden", r.ForbiddenRoom)
	group.Register("message", r.RoomMessage)
	group.Register("roommate", r.RoomMate)
	group.Register("create", r.CreateRoom)
	group.Register("nat", r.Nat)
}
