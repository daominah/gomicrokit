package gofast

import (
	"time"
)

// Cron should be inited and run by calling NewCron
type Cron struct {
	job       func()
	interval  time.Duration
	remainder time.Duration
	lastJob   time.Time
	ticker    *time.Ticker
}

// Cron periodically executes the job in a goroutine at specific time.
// Example: remainder = 7 hours, interval = 24 hours: the job will be executed
// at 7 a.m. everyday (executed time point accuracy is interval / 100).
// Any function can be wrap: job = func(){yourFunc(args...)}
func NewCron(job func(), interval time.Duration, remainder time.Duration) *Cron {
	// initialize cron obj
	c := &Cron{
		job:       job,
		interval:  interval,
		remainder: remainder,
	}
	c.lastJob = time.Now().Add(-c.remainder).Truncate(c.interval).Add(c.remainder)

	// run
	tick := 1 * time.Second
	if tick > c.interval/100 {
		tick = c.interval / 100
	}
	c.ticker = time.NewTicker(tick)
	go func() {
		for {
			now := time.Now()
			if now.Sub(c.lastJob) >= c.interval {
				go c.job()
				c.lastJob = now.Add(-c.remainder).Truncate(c.interval).
					Add(c.remainder)
			}
			// sleep for a tick duration
			<-c.ticker.C
		}
	}()
	return c
}

func (c *Cron) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
}
