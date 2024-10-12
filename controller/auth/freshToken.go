package controller

import (
	"ginWeb/config"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
	"github.com/gin-gonic/gin"
	"time"
)

type FreshToken struct {
}

func (receiver FreshToken) V1(c *gin.Context) {
	tokenStr, f := c.Get("token")
	if !f {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.NoToken, Message: "invalid token",
		})
		return
	}
	token, f := tokenStr.(*auth.Token)
	if !f {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.WrongToken, Message: "invalid token",
		})
		return
	}
	// 时间延后
	token.Expire = time.Now().Add(time.Second * time.Duration(config.Conf.Server.TokenExpire))
	s, err := token.Sign()
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.WrongToken, Message: "invalid token",
		})
		return
	}
	c.JSON(200, dataType.JsonRes{
		Data: s,
	})
}

func (receiver FreshToken) RegisterRoute(r string, g *gin.RouterGroup) {
	g.Handle("GET", r, receiver.V1)
}
