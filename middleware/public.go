package middleware

import (
	"ginWeb/service/scheduler"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
)

// PeriodType 限流器周期类型
type PeriodType int

const (
	MinuteP PeriodType = iota
	HourP
	DayP
	All
)

func (p PeriodType) String() string {
	switch p {
	case MinuteP:
		return "minute"
	case HourP:
		return "hour"
	case DayP:
		return "day"
	case All:
		return "all"
	default:
		return "unknown"
	}
}

type Middleware interface {
	Handle(ctx *gin.Context)
}

type Limiter interface {
	// Reset 周期性重置计数器
	Reset(PeriodType)
	Handle(*gin.Context)
}

// 所有已生成的限流器
var limiterContainer *[]Limiter

func init() {
	// 添加限流器reset的定时任务
	limiterContainer = &[]Limiter{}
	_, err := scheduler.App.AddFunc("0 * * * * *", func() {
		loguru.Logger.Debugf("limiter minute period resetting. limiter count: %d", len(*limiterContainer))
		for _, v := range *limiterContainer {
			go v.Reset(MinuteP)
		}
	})
	if err != nil {
		loguru.Logger.Fatal(err)
	}

	_, err = scheduler.App.AddFunc("0 0 * * * *", func() {
		loguru.Logger.Debugf("limiter hour period resetting. limiter count: %d", len(*limiterContainer))
		for _, v := range *limiterContainer {
			go v.Reset(HourP)
		}
	})
	if err != nil {
		loguru.Logger.Fatal(err)
	}

	_, err = scheduler.App.AddFunc("0 0 0 * * *", func() {
		loguru.Logger.Debugf("limiter day period resetting. limiter count: %d", len(*limiterContainer))
		for _, v := range *limiterContainer {
			go v.Reset(DayP)
		}
	})
	if err != nil {
		loguru.Logger.Fatal(err)
	}
}
