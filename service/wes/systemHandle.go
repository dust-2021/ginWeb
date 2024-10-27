package wes

import (
	"time"
)

// 响应心跳检测的响应
func heartbeatHandle(w *WContext) {
	w.Conn.heartChan <- time.Now().Unix()
}

func init() {
	RegisterHandler("heartbeat", heartbeatHandle)
}
