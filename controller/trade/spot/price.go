package spot

import (
	"ginWeb/service/dataType"
	"ginWeb/utils/exchange/binance"
	"ginWeb/utils/exchange/binance/binanceApi"
	"github.com/gin-gonic/gin"
)

type Price struct {
}

func (p Price) V1(c *gin.Context) {
	inter := binanceApi.SpotPrice{}
	ex := binance.Binance{}
	err := ex.SyncRequests(&inter)
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: err.Error(),
		})
		return
	}
	result, err := inter.GetResult()
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: err.Error(),
		})
		return
	}
	c.JSON(200, dataType.JsonRes{
		Data: result,
	})
}

func (p Price) RegisterRoute(r string, g *gin.RouterGroup) {
	g.Handle("GET", r, p.V1)
}
