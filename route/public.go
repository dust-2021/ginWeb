package route

import (
	"ginWeb/config"
	"ginWeb/controller/perm"
	"ginWeb/controller/user"
	"ginWeb/controller/ws"
	"ginWeb/middleware"
	"ginWeb/service/wes"
	"ginWeb/service/wes/subscribe"
	"github.com/gin-gonic/gin"
	"time"
)
import "ginWeb/controller/auth"

// InitRoute 注册路由函数
func InitRoute(g *gin.Engine) {

	// 开放api组
	api := g.Group("/api")
	controller.Login{}.RegisterRoute("/login", api)

	// 带token验证的api组
	sapi := g.Group("/sapi")
	sapi.Use(middleware.NewLoginStatus().HttpHandle)
	controller.Logout{}.RegisterRoute("/logout", sapi)
	controller.FreshToken{}.RegisterRoute("/freshToken", sapi)

	// 系统api组
	systemApi := sapi.Group("/system")
	perm.Grant{}.RegisterRoute("/grant", systemApi)
	user.Users{}.RegisterRoute("/user", systemApi)

}

// InitWs 注册ws处理逻辑
func InitWs(g *gin.Engine) {
	// websocket
	g.Handle("GET", "/ws",
		middleware.NewIpLimiter(10, 0, 0, "ws").HttpHandle,
		middleware.NewLoginStatus().HttpHandle,
		wes.UpgradeConn)

	baseGroup := wes.NewGroup("base")
	baseGroup.Register("hello", ws.Hello)
	baseGroup.Register("time", ws.ServerTime)

	channelGroup := wes.NewGroup("channel")
	channelGroup.Register("broadcast", middleware.NewLoginStatus().WsHandle,
		middleware.NewPermission([]string{"admin"}).WsHandle, ws.Broadcast)
	channelGroup.Register("subscribe", ws.SubHandle)
	channelGroup.Register("unsubscribe", ws.UnsubHandle)

	roomGroup := wes.NewGroup("room")
	roomGroup.Register("create", ws.CreateRoom)
	roomGroup.Register("in", ws.GetInRoom)
	roomGroup.Register("out", ws.GetOutRoom)
	roomGroup.Register("close", ws.CloseRoom)
	roomGroup.Register("mates", ws.RoomMate)
	roomGroup.Register("list", ws.ListRoom)

	// 注册订阅事件
	if config.Conf.Server.Debug {
		subscribe.NewPublisher("hello", "1s", func() string {
			return "hello"
		})
	}
	subscribe.NewPublisher("time", "*/10 * * * * *", func() string {
		return time.Now().Format("2006-01-02 15:04:05.0000")
	})

	// 注册系统大厅
	subscribe.NewPublisher("hall", "")
}
