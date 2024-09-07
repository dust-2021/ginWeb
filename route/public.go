package route

import "github.com/gin-gonic/gin"
import "ginWeb/controller/auth"

func InitRoute(g *gin.Engine) error {
	api := g.Group("/api")
	err := controller.Login{Url: "/login"}.RegisterRoute(api)
	return err
}
