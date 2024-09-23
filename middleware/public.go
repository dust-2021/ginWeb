package middleware

import (
	"ginWeb/service/scheduler"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
)

type Middleware interface {
	Handle(ctx *gin.Context)
}

type Limiter interface {
	Reset(PeriodType)
	Handle(*gin.Context)
}

var limiterContainer *[]Limiter

func init() {
	// 添加限流器reset的定时任务
	limiterContainer = &[]Limiter{}
	_, err := scheduler.ScheduleApp.AddFunc("0 * * * * *", func() {
		loguru.Logu.Debug("limiter minute period resetting.")
		for _, v := range *limiterContainer {
			go v.Reset(MinuteP)
		}
	})
	if err != nil {
		loguru.Logu.Fatal(err)
	}
}
