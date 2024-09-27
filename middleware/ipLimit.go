package middleware

import (
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ipLimiter struct {
	uuid     string
	minute   map[string]uint32
	hour     map[string]uint32
	day      map[string]uint32
	minuteLm uint32
	hourLm   uint32
	dayLm    uint32
}

func (i *ipLimiter) Reset(p PeriodType) {
}

func (i *ipLimiter) Handle(c *gin.Context) {

}

// NewIpLimiter ip限流器，数据缓存在redis中，可选一个string参数作为限流器唯一ID，否则使用UUID生成
func NewIpLimiter(minute uint32, hour uint32, day uint32, args ...string) Limiter {
	var u string
	if len(args) > 0 {
		u = args[0]
	} else {
		uid, err := uuid.NewUUID()
		if err != nil {
			loguru.Logger.Fatal("create ipLimiter uuid failed")
		}
		u = uid.String()
	}
	limiter := &ipLimiter{
		uuid:     u,
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
