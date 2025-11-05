package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"golangbreaker/golangbreaker"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	runtime.GOMAXPROCS(2)
	http.Handle("/metrics", promhttp.Handler())
	bench := NewBench()
	http.Handle("/", bench)
	addr := ":9100"
	fmt.Println("Starting server on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic("启动 HTTP 服务失败: " + err.Error())
	}
}

var requestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "demo_request_total",
		Help: "Total number of demo requests",
	},
	[]string{"method", "event"},
)

var durationTotal = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "demo_request_duration_seconds",
	},
	[]string{"method", "event"},
)

func init() {
	prometheus.MustRegister(requestTotal, durationTotal)
}

type Bench struct {
	method  string
	breaker *golangbreaker.GoSchedBreaker
}

func NewBench() *Bench {
	return &Bench{
		method:  "gosched",
		breaker: golangbreaker.NewGoSchedBreaker(),
	}
}

func (b *Bench) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result := b.Work()
	_, _ = w.Write([]byte(result))
}

func (b *Bench) Work() string {
	start := time.Now()
	result := b.work()
	requestTotal.WithLabelValues(b.method, result).Inc()
	durationTotal.WithLabelValues(b.method, result).Observe(time.Since(start).Seconds())
	return result
}

func (b *Bench) work() string {
	if b.breaker.Break() {
		return "FastReject"
	}
	resp, err := http.Get("https://www.baidu.com/")
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	return "OK"
}
