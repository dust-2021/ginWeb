package main

import (
	"fmt"
	"ginWeb/config"
	"ginWeb/model/inital"
	"ginWeb/service/scheduler"
	"ginWeb/utils/loguru"
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
	g := gin.New()
	g.Use(gin.Recovery())

	ginLogConf := gin.LoggerConfig{
		Output: loguru.Logu.Writer(),
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[GIN] | %3d | %13v | %15s | %-7s %#v\n%s",
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
				param.ErrorMessage,
			)
		},
	}
	g.Use(gin.LoggerWithConfig(ginLogConf))
	// 注册路由
	_ = route.InitRoute(g)
	return g
}

func main() {
	g := application()
	scheduler.ScheduleApp.Start()
	_ = g.Run(fmt.Sprintf(":%d", config.Conf.Server.Port))
}
