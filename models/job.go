package models

import "github.com/robfig/cron"

type Job struct {
	Cron *cron.Cron
	DB   *DatabaseConfig
}
