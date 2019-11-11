package main

import (
	"math/rand"
	"time"

	"github.com/daominah/gomicrokit/log"
)

func worker(workerId int, inChan chan *int, outChan chan int) {
	sum := 0
	for {
		data := <-inChan
		// a receive from a closed channel returns the zero value immediately
		if data == nil {
			break
		}
		log.Debugf("worker %v received data: %v", workerId, *data)
		sum += *data
		_ = time.Sleep
		time.Sleep(time.Duration(rand.Int63n(int64(2 * time.Millisecond))))
	}
	log.Infof("worker %v is about to return: %v", workerId, sum)
	outChan <- sum
}

func main() {
	nWorkers := 4
	inChan := make(chan *int)
	outChan := make(chan int)
	for i := 0; i < nWorkers; i++ {
		go worker(i, inChan, outChan)
	}

	log.Infof("starting the main")
	for i := 0; i < 100; i++ {
		clonedValue := i
		inChan <- &clonedValue
	}
	close(inChan)
	sum := 0
	for i := 0; i < nWorkers; i++ {
		r := <-outChan
		sum += r
	}
	log.Infof("sum: %v", sum)
}
