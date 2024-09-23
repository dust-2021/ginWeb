package middleware

import (
	"github.com/gin-gonic/gin"
	"sync"
)

type ipLimiter struct {
	lock     sync.Mutex
	minute   map[string]uint32
	hour     map[string]uint32
	day      map[string]uint32
	minuteLm uint32
	hourLm   uint32
	dayLm    uint32
}

func (i *ipLimiter) Reset(p PeriodType) {
	i.lock.Lock()
	defer i.lock.Unlock()
}

func (i *ipLimiter) Handle(c *gin.Context) {
	i.lock.Lock()
	defer i.lock.Unlock()

}

func NewIpLimiter(minute uint32, hour uint32, day uint32) Limiter {
	limiter := &ipLimiter{
		minute:   make(map[string]uint32),
		hour:     make(map[string]uint32),
		day:      make(map[string]uint32),
		minuteLm: minute,
		hourLm:   hour,
		dayLm:    day,
	}
	*limiterContainer = append(*limiterContainer, limiter)
	return limiter
}
