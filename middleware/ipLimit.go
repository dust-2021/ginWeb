package middleware

import (
	"context"
	"fmt"
	reCache "ginWeb/service/cache"
	"ginWeb/service/dataType"
	"ginWeb/utils/database"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var namespace = "IPLimiter"

type ipLimiter struct {
	uuid     string
	minuteLm uint32
	hourLm   uint32
	dayLm    uint32
}

func (i *ipLimiter) key(p PeriodType, ip string) string {
	return fmt.Sprintf("%s::%s::%s", i.uuid, p.String(), ip)
}

func (i *ipLimiter) Reset(p PeriodType) {
	ctx := context.Background()
	resp := database.Rdb.Keys(ctx, fmt.Sprintf("::%s::%s::%s*", namespace, i.uuid, p.String()))
	if resp.Err() != nil {
		loguru.Logger.Error("ipLimiter reset err:" + resp.Err().Error())
	}
	keys, err := resp.Result()
	if err != nil {
		loguru.Logger.Error("ipLimiter reset err:" + err.Error())
	}
	for _, key := range keys {
		go database.Rdb.Del(ctx, key)
	}
}

func (i *ipLimiter) Handle(c *gin.Context) {
	ip := c.ClientIP()
	countM, err := reCache.Incr(namespace, i.key(MinuteP, ip))
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: "limiter failed",
		})
		return
	}
	countH, err := reCache.Incr(namespace, i.key(HourP, ip))
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: "limiter failed",
		})
		return
	}
	countD, err := reCache.Incr(namespace, i.key(DayP, ip))
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: 1, Message: "limiter failed",
		})
		return
	}
	if (i.minuteLm > 0 && countM > int64(i.minuteLm)) || (i.hourLm > 0 && countH > int64(i.hourLm)) || (i.dayLm > 0 && countD > int64(i.dayLm)) {
		c.AbortWithStatusJSON(403, dataType.JsonWrong{
			Code: 1, Message: "denied by IPLimiter",
		})
		return
	}
}

// NewIpLimiter ip限流器，数据缓存在redis中，可选一个string参数作为限流器唯一ID（尽量选接口路由），否则使用UUID生成
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
		minuteLm: minute,
		hourLm:   hour,
		dayLm:    day,
	}
	*limiterContainer = append(*limiterContainer, limiter)
	return limiter
}
