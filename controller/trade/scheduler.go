package trade

import (
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"ginWeb/service/exchangeCore"
	"github.com/gin-gonic/gin"
)

type ScheController struct {
}

func (s ScheController) Start(c *gin.Context) {
	exchangeCore.ExchangeSche.Start()
	c.JSON(200, dataType.JsonRes{
		Code: 0, Data: "success",
	})
}
func (s ScheController) Stop(c *gin.Context) {
	exchangeCore.ExchangeSche.Stop()
	c.JSON(200, dataType.JsonRes{
		Code: 0, Data: "success",
	})
}

func (s ScheController) RegisterRoute(r string, g *gin.RouterGroup) {
	g.Use(middleware.Permission{SelectPermission: [][]string{{"admin", "exchangeAdmin"}}}.Handle)
	g.Handle("GET", r+"/start", s.Start)
	g.Handle("GET", r+"/stop", s.Stop)
}
