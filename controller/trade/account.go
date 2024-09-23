package trade

import (
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"ginWeb/utils/exchange/binance"
	"ginWeb/utils/exchange/binance/binanceApi"
	"github.com/gin-gonic/gin"
)

type Account struct {
}

type addData struct {
	ApiKey string `json:"apiKey"`
	Secret string `json:"secret"`
}

func (a Account) AddV1(c *gin.Context) {
	data := addData{}
	err := c.Bind(&data)
	if err != nil {
		c.JSON(200, dataType.JsonWrong{
			Code: 1, Message: err.Error(),
		})
		return
	}
	// 检查权限
	ex := binance.Binance{
		ApiKey:     data.ApiKey,
		Secret:     data.Secret,
		RecvWindow: 5000,
	}
	checker := binanceApi.KeyCheck{}
	ex.Request(&checker)
	f, err := checker.GetResult()
	if err != nil {
		c.JSON(200, dataType.JsonWrong{
			Code: 1, Message: err.Error(),
		})
		return
	}
	if !f {
		c.JSON(200, dataType.JsonWrong{
			Code: 1, Message: "check failed",
		})
		return
	}

	c.JSON(200, dataType.JsonRes{})

}

func (a Account) RegisterRoute(r string, g *gin.RouterGroup) {
	g.Handle("POST", r, middleware.NewPermission([]string{"admin"}).Handle, a.AddV1)
}
