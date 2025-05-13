package server

import (
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"ginWeb/utils/auth"
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

func (s Server) Check(c *gin.Context) {
	key := c.Query("key")
	c.AbortWithStatusJSON(200, dataType.JsonRes{
		Code: dataType.Success,
		Data: auth.HashString("mole", key),
	})
}

func (s Server) RegisterRoute(r string, g *gin.RouterGroup) {
	g.Handle("GET", r+"/time", middleware.NewIpLimiter(120, 0, 0, g.BasePath()+r+"/time").HttpHandle, s.time)
	g.Handle("GET", r+"/check", middleware.NewIpLimiter(120, 0, 0, g.BasePath()+r+"/check").HttpHandle, s.Check)
}
