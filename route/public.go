package route

import (
	"ginWeb/config"
	configApi "ginWeb/controller/config"
	"ginWeb/controller/debug"
	"ginWeb/controller/perm"
	"ginWeb/controller/server"
	"ginWeb/controller/user"
	"ginWeb/controller/ws"
	"ginWeb/middleware"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
	"time"

	"github.com/gin-gonic/gin"

	controller "ginWeb/controller/auth"
)

// InitRoute 注册路由函数
func InitRoute(g *gin.Engine) {

	// 开放api组
	api := g.Group("/api")
	api.Handle("POST", "/register", middleware.NewIpLimiter(10, 0, 0, "/register").HttpHandle, user.PublicRegister)
	controller.Login{}.RegisterRoute("/login", api)
	server.Server{}.RegisterRoute("/server", api)

	// 带token验证的api组
	sapi := g.Group("/sapi")
	sapi.Use(middleware.AuthMiddle.HttpHandle)
	configApi.Admin{}.RegisterRoute("/config", sapi)
	controller.Logout{}.RegisterRoute("/logout", sapi)
	controller.FreshToken{}.RegisterRoute("/freshToken", sapi)
	server.InfoMessage{}.RegisterRoute("/info", sapi)

	// 系统api组
	systemApi := sapi.Group("/system")
	perm.Grant{}.RegisterRoute("/grant", systemApi)
	user.Users{}.RegisterRoute("/user", systemApi)
	perm.Permission{}.RegisterRoute("/permission", systemApi)

	if config.Conf.Server.Debug {
		debugGroup := g.Group("/debug")
		debug.WgDebug{}.RegisterRoute("/wg", debugGroup)
	}
}

// InitWs 注册ws处理逻辑
func InitWs(g *gin.Engine) {
	defer wes.PrintTasks()
	// websocket
	g.Handle("GET", "/ws",
		middleware.NewIpLimiter(10, 0, 0, "ws").HttpHandle,
		wes.UpgradeConn)
	// 用于ws提供的某些http接口
	wsApi := g.Group("/ws")
	wsApi.Use(middleware.AuthMiddle.HttpHandle)

	base := ws.Base{}
	base.RegisterWSRoute("base", wes.BasicGroup)

	channel := ws.ChannelController{}
	channel.RegisterWSRoute("channel", wes.BasicGroup)

	room := ws.RoomController{}
	room.RegisterWSRoute("room", wes.BasicGroup)
	room.RegisterRoute("room", wsApi)

	// 注册订阅事件
	if config.Conf.Server.Debug {
		subscribe.Publishers.NewPublisher("hello", "1s", func() string {
			return "hello"
		})
	}
	subscribe.Publishers.NewPublisher("time", "*/10 * * * * *", func() string {
		return time.Now().Format("2006-01-02 15:04:05.000")
	})

	// 注册系统大厅
	subscribe.Publishers.NewPublisher("hall", "")
}
