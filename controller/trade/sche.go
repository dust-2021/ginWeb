package trade

import (
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"github.com/gin-gonic/gin"
)

type TaskController struct {
}

func (s TaskController) Start(c *gin.Context) {
	//exchangeCore.ExchangeSche.Start()
	c.JSON(200, dataType.JsonRes{
		Code: 0, Data: "success",
	})
}
func (s TaskController) Stop(c *gin.Context) {
	//exchangeCore.ExchangeSche.Stop()
	c.JSON(200, dataType.JsonRes{
		Code: 0, Data: "success",
	})
}

func (s TaskController) RegisterRoute(r string, g *gin.RouterGroup) {
	g.Use(middleware.NewPermission([]string{}, []string{"admin", "exchangeAdmin"}).Handle)
	g.Handle("GET", r+"/start", s.Start)
	g.Handle("GET", r+"/stop", s.Stop)
}
