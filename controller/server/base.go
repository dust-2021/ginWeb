package server

import (
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"github.com/gin-gonic/gin"
	"time"
)

type Server struct{}

func (s Server) time(c *gin.Context) {
	var t = time.Now().UnixMilli()
	c.AbortWithStatusJSON(200, dataType.JsonRes{
		Code: dataType.Success,
		Data: t,
	})
}

func (s Server) RegisterRoute(r string, g *gin.RouterGroup) {
	g.Handle("GET", r+"/time", middleware.NewIpLimiter(120, 0, 0, g.BasePath()+r+"/time").HttpHandle, s.time)
}
