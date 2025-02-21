package ws

import (
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"time"
)

func Hello(w *wes.WContext) {
	w.Result(dataType.Success, "hello")
}

func ServerTime(w *wes.WContext) {
	w.Result(dataType.Success, time.Now().UnixMilli())
}
