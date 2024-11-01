package ws

import (
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
)

func CreateRoom(w *wes.WContext) {
	room, err := subscribe.NewRoom(w.Conn)
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
