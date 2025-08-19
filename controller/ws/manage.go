package ws

import (
	"ginWeb/config"
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Manage struct {
}

func (m Manage) WsConfigs(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusOK, dataType.JsonRes{
		Code: dataType.Success, Data: config.Conf.Server,
	})
}

func (m Manage) RegisterRoute(r string, g *gin.RouterGroup) {
	router := g.Group(r)
	router.Use(middleware.NewPermission([]string{"admin"}).HttpHandle)

	router.Handle("GET", "")
}
