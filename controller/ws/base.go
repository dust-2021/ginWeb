package ws

import (
	"ginWeb/model"
	"ginWeb/model/systemMode"
	reCache "ginWeb/service/cache"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"
	"strconv"
	"time"
)

func Hello(w *wes.WContext) {
	w.Result(dataType.Success, "hello")
}

func ServerTime(w *wes.WContext) {
	w.Result(dataType.Success, time.Now().UnixMilli())
}

// Login 保持ws登录状态，时间根据配置设定
func Login(w *wes.WContext) {
	if len(w.Request.Params) != 2 {
		w.Result(dataType.WrongBody, "invalid param")
		return
	}
	username := w.Request.Params[0]
	password := w.Request.Params[1]

	// 查找用户
	var record systemMode.User
	result := database.Db.Where("phone = ?", username).Or("email = ?", username).First(&record)
	pwd, err := auth.HashPassword(password)
	if result.Error != nil || err != nil || pwd != record.PasswordHash {
		w.Result(dataType.WrongData, "invalid username or password")
	}
	_, err = reCache.Get("wsOnline", strconv.FormatInt(record.Id, 10))
	if err != nil {
		w.Result(dataType.AlreadyExist, err.Error())
	}
	// 查询权限
	permissions, err := model.GetPermissionById(record.Id)
	if err != nil {
		w.Result(dataType.WrongBody, err.Error())
	}
	w.Conn.Login(record.Id, record.Email, permissions)
	w.Result(dataType.Success, "success")
}

func Logout(w *wes.WContext) {
	w.Conn.Logout()
	w.Result(dataType.Success, "success")
}
