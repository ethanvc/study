package main

import (
	"context"
	"errors"
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
	"github.com/redis/go-redis/v9"
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
	result := b.Work(r.Context())
	_, _ = w.Write([]byte(result))
}

func (b *Bench) Work(ctx context.Context) string {
	start := time.Now()
	result := b.work(ctx)
	requestTotal.WithLabelValues(b.breaker.Name(), result).Inc()
	durationTotal.WithLabelValues(b.breaker.Name()).Observe(time.Since(start).Seconds())
	return result
}

func (b *Bench) work(ctx context.Context) string {
	if b.breaker.Break() {
		return "FastReject"
	}
	return redisPayload(ctx)
}

func clearEnv(s string) {
	os.Unsetenv(s)
	os.Unsetenv(strings.ToLower(s))
	os.Unsetenv(strings.ToUpper(s))
}

func redisPayload(ctx context.Context) string {
	cmd := sRdb.Get(ctx, "key")
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return "OK"
		}
		return "Error"
	}
	return "OK"
}

var sRdb = redis.NewUniversalClient(&redis.UniversalOptions{
	Addrs: []string{"localhost:6379"}, // Redis 地址
	// 不设置密码
	DB:       0,  // 默认数据库
	PoolSize: 10, // 连接池大小
})
