package exchangeCore

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

var scheduler *cron.Cron

var priceLock sync.RWMutex

func GetSymbolPrice() {
	priceLock.RLock()
	defer priceLock.RUnlock()
	ticker := time.NewTicker(200 * time.Millisecond)
	for i := 0; i < 5; i++ {
		select {
		case <-ticker.C:
			fmt.Printf("\r %s 任务触发", time.Now().Format("2006-01-02 15:04:05"))
		}
	}
	ticker.Stop()
}

func init() {
	scheduler = cron.New(cron.WithSeconds())
	_, err := scheduler.AddFunc("* * * * * *", GetSymbolPrice)
	if err != nil {
		return
	}
}
