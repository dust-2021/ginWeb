package controller

import "github.com/gin-gonic/gin"

type BaseController interface {
	// RegisterRoute 注册接口
	RegisterRoute(g *gin.RouterGroup) error
}
