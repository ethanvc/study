package main

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"
)

func main() {
	runtime.GOMAXPROCS(2)
	limiter := NewLimiter()
	ch := make(chan int, 1000)
	var rejectCount atomic.Int64
	var processedCount atomic.Int64
	go func() {
		for _ = range ch {
			go func() {
				if limiter.Reject() {
					rejectCount.Add(1)
					return
				}
				time.Sleep(time.Millisecond * 100)
				processedCount.Add(1)
			}()
		}
	}()
	go func() {
		var lastRejectCount int64
		var lastProcessedCount int64
		for {
			time.Sleep(time.Second * 2)
			currentRejectCount := rejectCount.Load()
			currentProcessedCount := processedCount.Load()
			fmt.Printf("reject rejectCount:%d\n", currentRejectCount-lastRejectCount)
			fmt.Printf("processedCount:%d\n", currentProcessedCount-lastProcessedCount)
			lastRejectCount = currentRejectCount
			lastProcessedCount = processedCount.Load()
		}
	}()

	for {
		ch <- 2
	}
}

type Limiter struct {
	ratio atomic.Int64
}

func NewLimiter() *Limiter {
	l := &Limiter{}
	go l.run()
	return l
}

func (l *Limiter) Reject() bool {
	n := rand.Intn(10000)
	return n < int(l.ratio.Load())
}

func (l *Limiter) run() {
	const sleepDuration = 5
	for {
		start := time.Now()
		time.Sleep(sleepDuration)
		timeCost := time.Since(start)
		ratio := calcRatio(int(timeCost/time.Millisecond), sleepDuration, 1)
		l.ratio.Store(int64(ratio))
	}
}

func calcRatio(RT, T int, K float64) int {
	D := abs(RT - T)
	// 限制K的范围（0 < K ≤ 10）
	if K <= 0 {
		K = 0.1 // 避免K=0导致无变化
	} else if K > 10 {
		K = 10
	}
	// 核心计算
	x := (K * float64(D)) / 1000
	ratio := 10000 * (1 - math.Exp(-x))
	// 边界处理（确保在0-10000之间）
	if ratio < 0 {
		return 0
	}
	if ratio > 10000 {
		return 10000
	}
	return int(ratio)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
