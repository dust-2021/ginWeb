package main

import (
	"fmt"
	"ginWeb/config"
	"ginWeb/model/inital"
	_ "ginWeb/service/exchangeCore"
	"ginWeb/service/scheduler"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	_ "net/http/pprof"
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
	// 日志中间件
	ginLogConf := gin.LoggerConfig{
		Output: loguru.Logger.Writer(),
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
	route.InitRoute(g)
	route.InitWs(g)
	return g
}

func main() {

	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:8001", nil))
	}()

	g := application()
	// 启动定时器
	scheduler.App.Start()
	_ = g.Run(fmt.Sprintf(":%d", config.Conf.Server.Port))
}
