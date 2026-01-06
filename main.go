package main

import (
	"fmt"
	"ginWeb/config"
	"ginWeb/model/inital"
	"ginWeb/service/scheduler"
	"ginWeb/service/udp"
	"ginWeb/service/wireguard"
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

	if config.Conf.Server.Debug {
		go func() {
			log.Println(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", config.Conf.Server.PprofPort), nil))
		}()
	}
	g := application()
	// 启动定时器
	scheduler.App.Start()
	err := wireguard.WireguardManager.Start()
	if err != nil {
		panic(fmt.Sprintf("start wireguard failed: %s", err.Error()))
	}
	err = udp.UdpSvr.Run()
	if err != nil {
		panic(fmt.Sprintf("start udp server failed: %s", err.Error()))
	}
	_ = g.Run(fmt.Sprintf(":%d", config.Conf.Server.Port))
}
