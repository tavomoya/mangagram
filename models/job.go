package models

import "github.com/robfig/cron"

// Job is a struct that
// contains some properties that
// are used inside the CRON jobs
type Job struct {
	Cron *cron.Cron
	DB   *DatabaseConfig
}
