package ws

import (
	"fmt"
	"ginWeb/middleware/wsMiddle"
	"ginWeb/model"
	"ginWeb/model/systemMode"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/auth"
	"ginWeb/utils/database"
	"strings"
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
		w.Result(dataType.WrongData, "invalid username or password")
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

// ws订阅事件接口
func subHandle(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongData, "without params")
		return
	}
	var keys []string
	for _, v := range w.Request.Params {
		name, flag := v.(string)
		if !flag {
			w.Result(dataType.WrongData, fmt.Sprintf("param invalid %v", v))
			return
		}
		keys = append(keys, name)
	}
	var failedKeys = make([]string, 0)
	for _, name := range keys {
		if f := w.Conn.Subscribe(name); !f {
			failedKeys = append(failedKeys, name)
		}
	}
	if len(failedKeys) > 0 {
		w.Result(dataType.NotFound, strings.Join(failedKeys, ","))
	}
	w.Result(dataType.Success, "success")
}

// ws取消事件订阅接口
func unsubHandle(w *wes.WContext) {
	if len(w.Request.Params) == 0 {
		w.Result(dataType.WrongData, "without params")
		return
	}

	var keys []string
	for _, v := range w.Request.Params {
		name, flag := v.(string)
		if !flag {
			w.Result(dataType.WrongData, fmt.Sprintf("param invalid %v", v))
			return
		}
		keys = append(keys, name)
	}

	for _, name := range keys {
		w.Conn.UnSubscribe(name)
	}
	w.Result(dataType.Success, "success")
}

// Broadcast 向频道发送消息
func Broadcast(w *wes.WContext) {

	if len(w.Request.Params) != 2 {
		w.Result(dataType.WrongBody, "invalid param")
		return
	}
	name, ok := w.Request.Params[0].(string)
	if !ok {
		w.Result(dataType.WrongBody, "invalid channel")
		return
	}
	msg, ok := w.Request.Params[1].(string)
	if !ok {
		w.Result(dataType.WrongBody, "invalid message")
		return
	}
	err := w.Conn.Publish(name, msg)
	if err != nil {
		w.Result(dataType.WrongBody, err.Error())
		return
	}
	w.Result(dataType.Success, "success")
}

func init() {
	wes.RegisterHandler("login", Login)
	wes.RegisterHandler("time", ServerTime)
	wes.RegisterHandler("logout", wsMiddle.LoginCheck, Logout)
	wes.RegisterHandler("broadcast", Broadcast)
	wes.RegisterHandler("subscribe", subHandle)
	wes.RegisterHandler("unsubscribe", unsubHandle)
}
