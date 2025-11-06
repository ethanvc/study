package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"golangbreaker/golangbreaker"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	clearEnv("http_proxy")
	clearEnv("https_proxy")
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
	[]string{"method"},
)

func init() {
	prometheus.MustRegister(requestTotal, durationTotal)
}

type Bench struct {
	breaker golangbreaker.Breaker
}

func NewBench() *Bench {
	return &Bench{
		breaker: golangbreaker.NewGoSchedBreaker(),
	}
}

func (b *Bench) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.ReadAll(r.Body)
	result := b.Work()
	_, _ = w.Write([]byte(result))
}

func (b *Bench) Work() string {
	start := time.Now()
	result := b.work()
	requestTotal.WithLabelValues(b.breaker.Name(), result).Inc()
	durationTotal.WithLabelValues(b.breaker.Name()).Observe(time.Since(start).Seconds())
	return result
}

func (b *Bench) work() string {
	if b.breaker.Break() {
		return "FastReject"
	}
	resp, err := http.Get("https://www.baidu.com/")
	if err != nil {
		return "Error"
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return "OK"
}

func clearEnv(s string) {
	os.Unsetenv(s)
	os.Unsetenv(strings.ToLower(s))
	os.Unsetenv(strings.ToUpper(s))
}
