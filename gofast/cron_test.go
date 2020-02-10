package gofast

import (
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func TestCron(t *testing.T) {
	var lock sync.Mutex
	counter := make([]time.Time, 0)
	job := func() {
		lock.Lock()
		counter = append(counter, time.Now())
		lock.Unlock()
	}
	c := NewCron(job, 15*time.Millisecond, 1*time.Millisecond)
	time.Sleep(300 * time.Millisecond)
	c.Stop()
	time.Sleep(50 * time.Millisecond)
	_ = spew.Dump
	//spew.Dump(counter)
	if len(counter) != 20 {
		t.Error(len(counter))
	}
}
