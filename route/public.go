package route

import (
	"ginWeb/middleware"
	"github.com/gin-gonic/gin"
)
import "ginWeb/controller/auth"

func InitRoute(g *gin.Engine) error {

	api := g.Group("/api")
	controller.Login{}.RegisterRoute("/login", api)

	// 带token验证API
	sapi := g.Group("/sapi")
	sapi.Use(middleware.LoginStatus{}.Handle)
	controller.Logout{}.RegisterRoute("/logout", sapi)
	controller.FreshToken{}.RegisterRoute("/freshToken", sapi)
	return nil
}
