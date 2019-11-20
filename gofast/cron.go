package gofast

import (
	"time"

)

type Cron struct {
	job func()
	// ex: remainder = 7 hours, interval = 24 hours: the job will be executed
	// at 7 a.m. everyday
	interval  time.Duration
	remainder time.Duration
	lastJob   time.Time
	ticker    *time.Ticker
}

// periodically execute the job in a goroutine at specific time
func NewCron(job func(), interval time.Duration, remainder time.Duration) *Cron {
	c := &Cron{
		job:       job,
		interval:  interval,
		remainder: remainder,
	}
	c.lastJob = time.Now().Add(-c.remainder).Truncate(interval).Add(c.remainder)
	tick := 1 * time.Second
	if tick > interval/100 {
		tick = interval / 100
	}
	c.ticker = time.NewTicker(tick)
	go func() {
		for {
			<-c.ticker.C
			now := time.Now()
			if now.Sub(c.lastJob) < c.interval {
				continue
			}
			go job()
			//log.Debugf("execute a job %#v at %v", job, now.Format(time.RFC3339Nano))
			c.lastJob = now.Add(-c.remainder).Truncate(interval).Add(c.remainder)
		}
	}()
	return c
}

func (c *Cron) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
}
