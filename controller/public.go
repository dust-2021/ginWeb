package controller

import (
	"ginWeb/service/wes"
	"github.com/gin-gonic/gin"
)

type BaseController interface {
	// RegisterRoute 注册接口
	RegisterRoute(r string, g *gin.RouterGroup)
}

type WsController interface {
	RegisterWSRoute(r string, g *wes.Group)
}

type RestfulController interface {
	Get(*gin.Context)
	Post(*gin.Context)
	Put(*gin.Context)
	Delete(*gin.Context)
	Update(*gin.Context)
	RegisterRoute(r string, g *gin.RouterGroup)
}
