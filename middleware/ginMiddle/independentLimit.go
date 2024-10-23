package ginMiddle

import (
	"fmt"
	"ginWeb/middleware"
	"ginWeb/service/dataType"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync/atomic"
)

type independentLimiter struct {
	minute   uint32
	hour     uint32
	day      uint32
	MinuteLm uint32
	HourLm   uint32
	DayLm    uint32
}

func (r *independentLimiter) Reset(p middleware.PeriodType) {
	switch p {
	case middleware.MinuteP:
		atomic.StoreUint32(&r.minute, 0)
	case middleware.HourP:
		atomic.StoreUint32(&r.hour, 0)
	case middleware.DayP:
		atomic.StoreUint32(&r.day, 0)
	case middleware.All:
		atomic.StoreUint32(&r.minute, 0)
		atomic.StoreUint32(&r.hour, 0)
		atomic.StoreUint32(&r.day, 0)
	default:
		loguru.SimpleLog(loguru.Error, "LIMITER", fmt.Sprintf("unknown period type %d", p))
	}

}

func (r *independentLimiter) Handle(c *gin.Context) {
	if r.MinuteLm != 0 && atomic.LoadUint32(&r.minute) > r.MinuteLm {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, dataType.JsonWrong{
			Code: dataType.RouteLimited, Message: fmt.Sprintf("server is busy"),
		})
		return
	}
	if r.HourLm != 0 && atomic.LoadUint32(&r.hour) > r.HourLm {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, dataType.JsonWrong{
			Code: dataType.RouteLimited, Message: fmt.Sprintf("server is busy"),
		})
		return
	}
	if r.DayLm != 0 && atomic.LoadUint32(&r.day) > r.DayLm {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, dataType.JsonWrong{
			Code: dataType.RouteLimited, Message: fmt.Sprintf("server is busy"),
		})
		return
	}
	atomic.AddUint32(&r.minute, 1)
	atomic.AddUint32(&r.hour, 1)
	atomic.AddUint32(&r.day, 1)

}

type rollingIndependentLimiter struct {
	count          uint32
	ReduceInMinute uint32
	Limit          uint32
}

func (r *rollingIndependentLimiter) Reset(p middleware.PeriodType) {
	if p == middleware.MinuteP || p == middleware.All {

		// 不必验证修改期间是否被修改
		//for {
		//	old := atomic.LoadUint32(&r.count)
		//	if old <= r.ReduceInMinute {
		//		atomic.StoreUint32(&r.count, 0)
		//		return
		//	}
		//	val := old - r.ReduceInMinute
		//	if atomic.CompareAndSwapUint32(&r.count, old, val) {
		//		return
		//	}
		//}
		old := atomic.LoadUint32(&r.count)
		if old <= r.ReduceInMinute {
			atomic.StoreUint32(&r.count, 0)
		} else {
			atomic.StoreUint32(&r.count, old-r.ReduceInMinute)
		}
	}
}

func (r *rollingIndependentLimiter) Handle(c *gin.Context) {
	if atomic.LoadUint32(&r.count) > r.Limit {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, dataType.JsonWrong{
			Code: dataType.RouteLimited, Message: "too much request in a time window",
		})
	}
}

// NewIndependentLimiter 定时reset的限流器，不同实例间单独计算
func NewIndependentLimiter(minute uint32, hour uint32, day uint32) middleware.Limiter {
	limiter := &independentLimiter{
		MinuteLm: minute,
		HourLm:   hour,
		DayLm:    day,
	}
	*middleware.LimiterContainer = append(*middleware.LimiterContainer, limiter)
	return limiter
}

// NewRollingIndependentLimiter 每分钟减少计数的限流器，不同实例单独计算
func NewRollingIndependentLimiter(minuteReduce uint32, limit uint32) middleware.Limiter {
	limiter := &rollingIndependentLimiter{
		ReduceInMinute: minuteReduce,
		Limit:          limit,
	}
	*middleware.LimiterContainer = append(*middleware.LimiterContainer, limiter)
	return limiter
}
