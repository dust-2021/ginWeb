package route

import (
	"ginWeb/config"
	"ginWeb/controller/perm"
	"ginWeb/controller/trade"
	"ginWeb/controller/trade/spot"
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

	// 交易相关api组
	tradeApi := sapi.Group("/trade")
	trade.Account{}.RegisterRoute("/account", tradeApi)

	// binance接口
	binance := tradeApi.Group("/binance")
	spot.Price{}.RegisterRoute("/api/v3/ticker/price", binance)
}

// InitWs 注册ws处理逻辑
func InitWs(g *gin.Engine) {
	// websocket
	g.Handle("GET", "/ws",
		middleware.NewIpLimiter(10, 0, 0, "ws").HttpHandle, wes.UpgradeConn)

	baseGroup := wes.NewGroup("base")
	baseGroup.Register("hello", ws.Hello)
	baseGroup.Register("login", ws.Login)
	baseGroup.Register("time", ws.ServerTime)
	baseGroup.Register("logout", middleware.NewLoginStatus().WsHandle, ws.Logout)

	channelGroup := wes.NewGroup("channel")
	channelGroup.Register("broadcast", middleware.NewLoginStatus().WsHandle,
		middleware.NewPermission([]string{"admin"}).WsHandle, ws.Broadcast)
	channelGroup.Register("subscribe", ws.SubHandle)
	channelGroup.Register("unsubscribe", ws.UnsubHandle)

	roomGroup := wes.NewGroup("room")
	roomGroup.Use(middleware.NewLoginStatus().WsHandle)
	roomGroup.Register("create", ws.CreateRoom)
	roomGroup.Register("in", ws.GetInRoom)
	roomGroup.Register("out", ws.GetOutRoom)
	roomGroup.Register("close", ws.CloseRoom)
	roomGroup.Register("mates", ws.RoomMate)

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
