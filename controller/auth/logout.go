package controller

import (
	reCache "ginWeb/service/cache"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"github.com/gin-gonic/gin"
	"time"
)

type Logout struct {
}

func (receiver Logout) V1(c *gin.Context) {
	tokenC, f := c.Get("token")
	token, ok := tokenC.(auth.Token)

	// 拉黑该token
	if f && ok {
		sign, _ := token.Sign()
		err := reCache.Set("blackToken", sign, 1, uint(token.Expire.Sub(time.Now()).Seconds()))
		if err != nil {
			return
		}
	}
	c.JSON(200, dataType.JsonRes{
		Code: 0, Data: "success",
	})
}

func (receiver Logout) RegisterRoute(r string, g *gin.RouterGroup) {
	g.Handle("GET", r, receiver.V1)
}
