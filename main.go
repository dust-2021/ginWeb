package main

import (
	"fmt"
	"ginWeb/config"
	"ginWeb/model/inital"
	"github.com/gin-gonic/gin"
)
import "ginWeb/route"

func application() *gin.Engine {
	if !config.Conf.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	// 初始化数据表
	if config.Conf.Database.Initial {
		inital.InitializeMode()
	}
	g := gin.Default()
	// 注册路由
	_ = route.InitRoute(g)
	return g
}

func main() {
	g := application()
	_ = g.Run(fmt.Sprintf(":%d", config.Conf.Server.Port))
}
