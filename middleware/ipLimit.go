package middleware

import (
	"context"
	"fmt"
	reCache "ginWeb/service/cache"
	"ginWeb/service/dataType"
	"ginWeb/service/wes"
	"ginWeb/utils/database"
	"ginWeb/utils/loguru"
	"github.com/gin-gonic/gin"
)

// redis中key的前缀
var namespace = "IPLimiter"

// 已存在的limiter的uid
var existedUuid []string

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
	resp := database.Rdb.Keys(ctx, fmt.Sprintf("::%s::%s::%s::*", namespace, i.uuid, p.String()))
	if resp.Err() != nil {
		loguru.SimpleLog(loguru.Error, "LIMITER", "ipLimiter reset err:"+resp.Err().Error())
	}
	keys, err := resp.Result()
	if err != nil {
		loguru.SimpleLog(loguru.Error, "LIMITER", "ipLimiter reset err:"+err.Error())
	}
	for _, key := range keys {
		go database.Rdb.Del(ctx, key)
	}
}

func (i *ipLimiter) handle(ip string) (error, bool) {
	countM, err := reCache.Incr(namespace, i.key(MinuteP, ip))
	if err != nil {
		return err, false
	}
	countH, err := reCache.Incr(namespace, i.key(HourP, ip))
	if err != nil {
		return err, false
	}
	countD, err := reCache.Incr(namespace, i.key(DayP, ip))
	if err != nil {
		return err, false
	}
	if (i.minuteLm > 0 && countM > int64(i.minuteLm)) || (i.hourLm > 0 && countH > int64(i.hourLm)) || (i.dayLm > 0 && countD > int64(i.dayLm)) {
		return nil, false
	}
	return nil, true
}

func (i *ipLimiter) HttpHandle(c *gin.Context) {
	err, ok := i.handle(c.ClientIP())
	if err != nil {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.Unknown, Message: "failed",
		})
		return
	}
	if !ok {
		c.AbortWithStatusJSON(200, dataType.JsonWrong{
			Code: dataType.IpLimited, Message: "denied",
		})
		return
	}
}

func (i *ipLimiter) WsHandle(c *wes.WContext) {
	err, ok := i.handle(c.Conn.IP)
	if err != nil {
		c.Result(dataType.Unknown, "denied")
		return
	}
	if !ok {
		c.Result(dataType.IpLimited, "denied")
		return
	}
}

// NewIpLimiter ip限流器，数据缓存在redis中
func NewIpLimiter(minute uint32, hour uint32, day uint32, uniqueId string) Limiter {
	// 判断是否已存在同名uuid
	for _, v := range existedUuid {
		if v == uniqueId {
			loguru.SimpleLog(loguru.Fatal, "LIMITER", "duplicate ipLimiter uuid:"+v)
		}
	}
	limiter := &ipLimiter{
		uuid:     uniqueId,
		minuteLm: minute,
		hourLm:   hour,
		dayLm:    day,
	}
	*LimiterContainer = append(*LimiterContainer, limiter)
	return limiter
}

func init() {
	existedUuid = []string{}
}
