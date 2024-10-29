package route

import (
	"ginWeb/controller/perm"
	"ginWeb/controller/trade"
	"ginWeb/controller/trade/spot"
	"ginWeb/controller/ws"
	"ginWeb/middleware/ginMiddle"
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
	wes.RegisterHandler("hello", ws.Hello)

	// 注册订阅事件
	subscribe.NewPublisher("hello", "1s", func() string {
		return "hello"
	})
	subscribe.NewPublisher("time", "*/2 * * * * *", func() string {
		return time.Now().Format("2006-01-02 15:04:05.0000")
	})

	// 注册频道
	subscribe.NewPublisher("hall", "")
}
