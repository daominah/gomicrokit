package gofast

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func TestCron(t *testing.T) {
	counter := make([]time.Time, 0)
	job := func() {
		counter = append(counter, time.Now())
	}
	c := NewCron(job, 5*time.Millisecond, 1*time.Millisecond)
	time.Sleep(300 * time.Millisecond)
	c.Stop()
	_ = spew.Dump
	//spew.Dump(counter)
	if len(counter) != 60 {
		t.Error(len(counter))
	}
}
