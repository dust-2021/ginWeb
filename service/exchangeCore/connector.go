package exchangeCore

import (
	reCache "ginWeb/service/cache"
	"ginWeb/utils/loguru"
	"github.com/robfig/cron/v3"
	"time"
)

var ExchangeSche *cron.Cron

func GetSymbolPrice() {
	ticker := time.NewTicker(200 * time.Millisecond)
	loguru.Logu.Infof("getting price")
	for i := 0; i < 5; i++ {
		select {
		case <-ticker.C:
			err := reCache.Set("exchange", "price", 0, 60)
			if err != nil {
				loguru.Logu.Errorf("get price failed")
			}
		}
	}
	ticker.Stop()
}

func init() {
	ExchangeSche = cron.New(cron.WithSeconds())
	_, _ = ExchangeSche.AddFunc("* * * * * *", GetSymbolPrice)
}
