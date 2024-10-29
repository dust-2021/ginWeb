package wsMiddle

import (
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
)

// LoginCheck ws登录验证中间件
func LoginCheck(w *wes.WContext) {
	if !w.Conn.LoginStatus() {
		w.Result(dataType.NoToken, "haven't login")
	}
}
