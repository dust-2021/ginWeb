package route

import (
	"ginWeb/controller/perm"
	"ginWeb/controller/trade"
	"ginWeb/controller/trade/spot"
	"ginWeb/controller/ws"
	"ginWeb/middleware/ginMiddle"
	"ginWeb/middleware/wsMiddle"
	"ginWeb/service/wes"
	"github.com/gin-gonic/gin"
)
import "ginWeb/controller/auth"

// InitRoute 注册路由函数
func InitRoute(g *gin.Engine) {

	// 开放api组
	api := g.Group("/api")
	controller.Login{}.RegisterRoute("/login", api)

	// 带token验证的api组
	sapi := g.Group("/sapi")
	sapi.Use(ginMiddle.NewLoginStatus().Handle)
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
	g.Handle("GET", "/ws", ginMiddle.NewIpLimiter(10, 0, 0, "ws").Handle, wes.UpgradeConn)
	wes.RegisterHandler("login", ws.Login)
	wes.RegisterHandler("hello", ws.Hello)
	wes.RegisterHandler("refresh", ws.Refresh)
	wes.RegisterHandler("time", ws.ServerTime)
	wes.RegisterHandler("logout", wsMiddle.LoginCheck, ws.Logout)
}
