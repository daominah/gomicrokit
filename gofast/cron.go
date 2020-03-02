package gofast

import (
	"context"
	"time"
)

// Cron must be initialized and run by calling NewCron
type Cron struct {
	job         func()
	interval    time.Duration
	remainder   time.Duration
	lastJob     time.Time
	ticker      *time.Ticker
	stopChan    <-chan struct{}
	stopChanCxl context.CancelFunc
}

// NewCron periodically executes input job in a goroutine at specific time.
// Example: remainder = 7 hours, interval = 24 hours: the job will be executed
// at 7 a.m. everyday (executed time point accuracy is interval / 100).
// Any function can be wrap: job = func(){yourFunc(args...)}
func NewCron(job func(), interval time.Duration, remainder time.Duration) *Cron {
	// initialize cron obj
	ctx, cxl := context.WithCancel(context.Background())
	c := &Cron{
		job:         job,
		interval:    interval,
		remainder:   remainder,
		stopChan:    ctx.Done(),
		stopChanCxl: cxl,
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
			select {
			case <-c.ticker.C:
				continue
			case <-c.stopChan:
				return
			}
		}
	}()
	return c
}

// Stop stops the cron's loop
func (c *Cron) Stop() {
	c.ticker.Stop()
	c.stopChanCxl()
}
