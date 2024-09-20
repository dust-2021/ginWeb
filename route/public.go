package route

import (
	"ginWeb/controller/trade"
	"ginWeb/middleware"
	"github.com/gin-gonic/gin"
)
import "ginWeb/controller/auth"

// InitRoute 注册路由函数
func InitRoute(g *gin.Engine) error {

	api := g.Group("/api")
	controller.Login{}.RegisterRoute("/login", api)

	// 带token验证API
	sapi := g.Group("/sapi")
	sapi.Use(middleware.LoginStatus{}.Handle)
	controller.Logout{}.RegisterRoute("/logout", sapi)
	controller.FreshToken{}.RegisterRoute("/freshToken", sapi)

	tradeApi := sapi.Group("/trade")

	trade.Account{}.RegisterRoute("/account", tradeApi)
	return nil
}
