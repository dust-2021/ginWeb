package ws

import (
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
)

func CreateRoom(w *wes.WContext) {
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

func GetInRoom(w *wes.WContext) {
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

func GetOutRoom(w *wes.WContext) {
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

func CloseRoom(w *wes.WContext) {
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

func RoomMate(w *wes.WContext) {
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

func ListRoom(w *wes.WContext) {
	w.Result(dataType.Success, subscribe.RoomInfo())
}
