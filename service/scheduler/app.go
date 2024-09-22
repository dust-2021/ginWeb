package scheduler

import "github.com/robfig/cron/v3"

var ScheduleApp *cron.Cron

func init() {
	ScheduleApp = cron.New(cron.WithSeconds())
}
