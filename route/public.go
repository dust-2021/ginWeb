package route

import (
	"ginWeb/controller/perm"
	"ginWeb/controller/trade"
	"ginWeb/controller/trade/spot"
	"ginWeb/middleware"
	"ginWeb/service/wes"
	"github.com/gin-gonic/gin"
)
import "ginWeb/controller/auth"

// InitRoute 注册路由函数
func InitRoute(g *gin.Engine) error {
	// websocket
	g.Handle("GET", "/ws", middleware.NewLoginStatus().Handle, wes.UpgradeConn)

	// 开放api组
	api := g.Group("/api")
	controller.Login{}.RegisterRoute("/login", api)

	// 带token验证的api组
	sapi := g.Group("/sapi")
	sapi.Use(middleware.NewLoginStatus().Handle)
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
	return nil
}

func InitWs() {

}
