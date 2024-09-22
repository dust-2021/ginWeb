package middleware

import (
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/service/scheduler"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync/atomic"
)

type PeriodType int

const (
	MinuteP PeriodType = iota
	HourP
	DayP
	All
)

type Limiter interface {
	Reset(PeriodType)
	Handle(*gin.Context)
}

var LimiterContainer *[]Limiter

// RouteLimiter 绑定在指定接口的限流器
type RouteLimiter struct {
	minute   uint32
	hour     uint32
	day      uint32
	MinuteLm uint32
	HourLm   uint32
	DayLm    uint32
}

func (r *RouteLimiter) Reset(p PeriodType) {
	switch p {
	case MinuteP:
		atomic.StoreUint32(&r.minute, 0)
	case HourP:
		atomic.StoreUint32(&r.hour, 0)
	case DayP:
		atomic.StoreUint32(&r.day, 0)
	case All:
		atomic.StoreUint32(&r.minute, 0)
		atomic.StoreUint32(&r.hour, 0)
		atomic.StoreUint32(&r.day, 0)
	default:
		loguru.Logu.Errorf("unknown period type %d", p)
	}

}

func (r *RouteLimiter) Handle(c *gin.Context) {
	if r.MinuteLm != 0 && atomic.LoadUint32(&r.minute) > r.MinuteLm {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, dataType.JsonWrong{
			Code: 1, Message: fmt.Sprintf("more than %d in a minute", r.MinuteLm),
		})
		return
	}
	if r.HourLm != 0 && atomic.LoadUint32(&r.hour) > r.HourLm {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, dataType.JsonWrong{
			Code: 1, Message: fmt.Sprintf("more than %d in a hour", r.HourLm),
		})
		return
	}
	if r.DayLm != 0 && atomic.LoadUint32(&r.day) > r.DayLm {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, dataType.JsonWrong{
			Code: 1, Message: fmt.Sprintf("more than %d in a day", r.DayLm),
		})
		return
	}
	atomic.AddUint32(&r.minute, 1)
	atomic.AddUint32(&r.hour, 1)
	atomic.AddUint32(&r.day, 1)

}

func NewRouteLimiter(minute uint32, hour uint32, day uint32) *RouteLimiter {
	limiter := &RouteLimiter{
		MinuteLm: minute,
		HourLm:   hour,
		DayLm:    day,
	}
	*LimiterContainer = append(*LimiterContainer, limiter)
	return limiter
}

func init() {
	LimiterContainer = &[]Limiter{}
	_, err := scheduler.ScheduleApp.AddFunc("0 * * * * *", func() {
		loguru.Logu.Debug("limiter minute period resetting.")
		for _, v := range *LimiterContainer {
			go v.Reset(MinuteP)
		}
	})
	if err != nil {
		loguru.Logu.Fatal(err)
	}
}
