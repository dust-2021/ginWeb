package config

import (
	"ginWeb/config"
	"ginWeb/middleware"
	"ginWeb/service/dataType"

	"github.com/gin-gonic/gin"
)

type Admin struct {
}

func (a Admin) SetPublicRegister(ctx *gin.Context) {
	v := ctx.Query("enable")
	switch v {
	case "true":
		config.DynamicConf.Set(config.EnablePublicRegister, true)
	case "false":
		config.DynamicConf.Set(config.EnablePublicRegister, false)
	default:
		ctx.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code:    dataType.WrongData,
			Message: "wrong parameter",
		})
		return
	}
	ctx.JSON(200, dataType.JsonRes{
		Code: dataType.Success,
		Data: config.DynamicConf.EnablePublicRegister(),
	})
}

func (a Admin) RegisterRoute(route string, g *gin.RouterGroup) {
	group := g.Group(route)
	group.Handle("GET", "/setPublicRegister", middleware.NewPermission([]string{"admin"}).HttpHandle, a.SetPublicRegister)
}
