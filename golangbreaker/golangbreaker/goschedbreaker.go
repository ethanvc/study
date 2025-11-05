package golangbreaker

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type GoSchedBreaker struct {
	overload atomic.Bool
}

func NewGoSchedBreaker() *GoSchedBreaker {
	breaker := &GoSchedBreaker{}
	go breaker.run()
	return breaker
}

func (b *GoSchedBreaker) Name() string {
	return "go_sched_breaker"
}

func (b *GoSchedBreaker) Break() bool {
	return b.overload.Load()
}

func (b *GoSchedBreaker) run() {
	for {
		overload := b.testOverload()
		b.overload.Store(overload)
		time.Sleep(50 * time.Millisecond)
	}
}

func (b *GoSchedBreaker) testOverload() bool {
	const routingCount = 3
	var wg sync.WaitGroup
	var seconds [routingCount]float64
	for i := 0; i < routingCount; i++ {
		wg.Go(func() {
			start := time.Now()
			runtime.Gosched()
			seconds[i] = time.Since(start).Seconds()
		})
	}
	wg.Wait()
	var avg float64
	for i := 0; i < routingCount; i++ {
		avg += seconds[i]
	}
	avg /= float64(routingCount)
	if avg > 0.001 {
		return true
	} else {
		return false
	}
}
