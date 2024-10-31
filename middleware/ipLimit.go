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
	"github.com/google/uuid"
)

var namespace = "IPLimiter"

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
	resp := database.Rdb.Keys(ctx, fmt.Sprintf("::%s::%s::%s*", namespace, i.uuid, p.String()))
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
	err, ok := i.handle(c.Conn.RemoteAddr().String())
	if err != nil {
		c.Result(dataType.Unknown, "denied")
		return
	}
	if !ok {
		c.Result(dataType.IpLimited, "denied")
		return
	}
}

// NewIpLimiter ip限流器，数据缓存在redis中，可选一个string参数作为限流器唯一ID，否则使用UUID生成
func NewIpLimiter(minute uint32, hour uint32, day uint32, args ...string) Limiter {
	var u string
	if len(args) > 0 {
		u = args[0]
	} else {
		uid, err := uuid.NewUUID()
		if err != nil {
			loguru.SimpleLog(loguru.Fatal, "LIMITER", "create ipLimiter uuid failed")
		}
		u = uid.String()
	}
	// 判断是否已存在同名uuid
	for _, v := range existedUuid {
		if v == u {
			loguru.SimpleLog(loguru.Fatal, "LIMITER", "duplicate ipLimiter uuid:"+v)
		}
	}
	limiter := &ipLimiter{
		uuid:     u,
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
