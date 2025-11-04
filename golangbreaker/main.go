package main

import (
	"golangbreaker/golangbreaker"
	"net/http"
	"net/http/httptest"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	runtime.GOMAXPROCS(2)
	http.Handle("/metrics", promhttp.Handler())
	go bench()
	err := http.ListenAndServe(":9100", nil)
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

func bench() {
	engine := gin.New()
	method := "test"
	breaker := golangbreaker.NewGoSchedBreaker()
	enableBreak := false
	engine.GET("/", func(c *gin.Context) {
		if enableBreak {
			if breaker.Break() {
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("FastReject"))
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("OK"))
	})
	for {
		originReq := httptest.NewRequest(http.MethodGet, "/", nil)
		go func() {
			start := time.Now()
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, originReq)
			requestTotal.WithLabelValues(method, recorder.Body.String()).Inc()
			durationTotal.WithLabelValues(method, recorder.Body.String()).Observe(time.Since(start).Seconds())
		}()
	}
}
