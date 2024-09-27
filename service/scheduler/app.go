package scheduler

import "github.com/robfig/cron/v3"

var App *cron.Cron

func init() {
	App = cron.New(cron.WithSeconds())
}
