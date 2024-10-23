package middleware

import (
	"fmt"
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

// LimiterContainer 所有已生成的限流器
var LimiterContainer *[]Limiter

func init() {
	// 添加限流器reset的定时任务
	LimiterContainer = &[]Limiter{}
	_, err := scheduler.App.AddFunc("0 * * * * *", func() {
		loguru.SimpleLog(loguru.Debug, "LIMITER", fmt.Sprintf("limiter minute period resetting. limiter count: %d", len(*LimiterContainer)))
		for _, v := range *LimiterContainer {
			go v.Reset(MinuteP)
		}
	})
	if err != nil {
		loguru.SimpleLog(loguru.Fatal, "LIMITER", err.Error())
	}

	_, err = scheduler.App.AddFunc("0 0 * * * *", func() {
		loguru.SimpleLog(loguru.Debug, "LIMITER", fmt.Sprintf("limiter hour period resetting. limiter count: %d", len(*LimiterContainer)))
		for _, v := range *LimiterContainer {
			go v.Reset(HourP)
		}
	})
	if err != nil {
		loguru.SimpleLog(loguru.Fatal, "LIMITER", err.Error())
	}

	_, err = scheduler.App.AddFunc("0 0 0 * * *", func() {
		loguru.SimpleLog(loguru.Debug, "LIMITER", fmt.Sprintf("limiter day period resetting. limiter count: %d", len(*LimiterContainer)))
		for _, v := range *LimiterContainer {
			go v.Reset(DayP)
		}
	})
	if err != nil {
		loguru.SimpleLog(loguru.Fatal, "LIMITER", err.Error())
	}
}
