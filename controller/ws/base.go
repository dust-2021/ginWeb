package ws

import (
	"ginWeb/model"
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"
	"time"
)

func Hello(w *wes.WContext) {
	w.Result(dataType.Success, "hello")
}

func Refresh(w *wes.WContext) {
	w.Conn.ResetLifeTime()
}

func ServerTime(w *wes.WContext) {

	w.Result(dataType.Success, time.Now().UnixMilli())
}

func Login(w *wes.WContext) {
	if len(w.Request.Params) != 2 {
		w.Result(dataType.WrongBody, "invalid param")
		return
	}
	username := w.Request.Params[0]
	password := w.Request.Params[1]
	u, ok := username.(string)
	if !ok {
		w.Result(dataType.WrongBody, "invalid param")
		return
	}
	p, ok := password.(string)
	if !ok {
		w.Result(dataType.WrongBody, "invalid param")
		return
	}
	// 查找用户
	var record systemMode.User
	result := database.Db.Where("phone = ?", u).Or("email = ?", u).First(&record)
	pwd, err := auth.HashPassword(p)
	if result.Error != nil || err != nil || pwd != record.PasswordHash {
		w.Result(dataType.WrongBody, "invalid username or password")
	}
	// 查询权限
	permissions, err := model.GetPermissionById(record.Id)
	if err != nil {
		w.Result(dataType.WrongBody, err.Error())
	}
	w.Conn.Login(record.Id, permissions)
	w.Result(dataType.Success, "success")
}

func Logout(w *wes.WContext) {
	w.Conn.Logout()
	w.Result(dataType.Success, "success")
}
