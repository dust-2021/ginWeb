package middleware

import (
	"fmt"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
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

func (i *independentLimiter) Reset(p PeriodType) {
	switch p {
	case MinuteP:
		atomic.StoreUint32(&i.minute, 0)
	case HourP:
		atomic.StoreUint32(&i.hour, 0)
	case DayP:
		atomic.StoreUint32(&i.day, 0)
	case All:
		atomic.StoreUint32(&i.minute, 0)
		atomic.StoreUint32(&i.hour, 0)
		atomic.StoreUint32(&i.day, 0)
	default:
		loguru.SimpleLog(loguru.Error, "LIMITER", fmt.Sprintf("unknown period type %d", p))
	}

}

func (i *independentLimiter) handle() bool {
	if i.MinuteLm != 0 && atomic.LoadUint32(&i.minute) > i.MinuteLm {
		return false
	}
	if i.HourLm != 0 && atomic.LoadUint32(&i.hour) > i.HourLm {
		return false
	}
	if i.DayLm != 0 && atomic.LoadUint32(&i.day) > i.DayLm {
		return false
	}
	atomic.AddUint32(&i.minute, 1)
	atomic.AddUint32(&i.hour, 1)
	atomic.AddUint32(&i.day, 1)
	return true
}

func (i *independentLimiter) HttpHandle(c *gin.Context) {
	if !i.handle() {
		c.AbortWithStatusJSON(http.StatusOK, dataType.JsonWrong{
			Code: dataType.TooManyRequests, Message: "denied",
		})
	}
}

func (i *independentLimiter) WsHandle(c *wes.WContext) {
	if !i.handle() {
		c.Result(dataType.TooManyRequests, "denied")
	}
}

type rollingIndependentLimiter struct {
	count          uint32
	ReduceInMinute uint32
	Limit          uint32
}

func (r *rollingIndependentLimiter) Reset(p PeriodType) {
	if p == MinuteP || p == All {

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

func (r *rollingIndependentLimiter) handle() bool {
	return atomic.LoadUint32(&r.count) < r.Limit
}

func (r *rollingIndependentLimiter) HttpHandle(c *gin.Context) {
	if !r.handle() {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, dataType.JsonWrong{
			Code: dataType.RouteLimited, Message: "denied",
		})
	}
}

func (r *rollingIndependentLimiter) WsHandle(c *wes.WContext) {
	if !r.handle() {
		c.Result(dataType.TooManyRequests, "denied")
	}
}

// NewIndependentLimiter 定时reset的限流器，不同实例间单独计算
func NewIndependentLimiter(minute uint32, hour uint32, day uint32) Limiter {
	limiter := &independentLimiter{
		MinuteLm: minute,
		HourLm:   hour,
		DayLm:    day,
	}
	*LimiterContainer = append(*LimiterContainer, limiter)
	return limiter
}

// NewRollingIndependentLimiter 每分钟减少计数的限流器，不同实例单独计算
func NewRollingIndependentLimiter(minuteReduce uint32, limit uint32) Limiter {
	limiter := &rollingIndependentLimiter{
		ReduceInMinute: minuteReduce,
		Limit:          limit,
	}
	*LimiterContainer = append(*LimiterContainer, limiter)
	return limiter
}
