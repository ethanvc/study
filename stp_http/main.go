package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func registerPprof(mux *http.ServeMux) {
	const prefix = "GET "
	mux.HandleFunc(prefix+"/debug/pprof/", pprof.Index)
	mux.HandleFunc(prefix+"/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc(prefix+"/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc(prefix+"/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc(prefix+"/debug/pprof/trace", pprof.Trace)
}

var (
	// 定义指标标签
	labels         = []string{"method", "event"}
	durationLabels = []string{"method"}

	// 请求计数器（用于计算QPS）
	requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "study_event_total",
		Help: "Total number of HTTP requests",
	}, labels)

	// 请求耗时分布（直方图）
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "study_event_duration_seconds",
		Help:    "HTTP request duration distribution",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 1.5, 3},
	}, durationLabels)
)

func unsetEnv(key string) {
	os.Unsetenv(key)
	os.Unsetenv(strings.ToUpper(key))
	os.Unsetenv(strings.ToLower(key))
}

func main() {
	const useH2C = true
	runtime.GOMAXPROCS(4)
	unsetEnv("http_proxy")
	unsetEnv("https_proxy")
	transport := http.DefaultTransport.(*http.Transport)
	transport.MaxIdleConnsPerHost = 100
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/echo", apiEchoHandler)
	mux.Handle("/metrics", promhttp.Handler())
	registerPprof(mux)
	var handler http.Handler
	handler = mux
	if useH2C {
		handler = h2c.NewHandler(handler, &http2.Server{})
	}
	httpSvr := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	const listenAddr = ":8080"
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	ln = &monitorListener{Listener: ln}
	fmt.Println("Server running on " + listenAddr)
	log.Fatal(httpSvr.Serve(ln))
}

type monitorListener struct {
	net.Listener
}

func (l *monitorListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err == nil {
		requestCounter.WithLabelValues(GlobalMethod, "NewIngressTcp").Inc()
	}
	return c, err
}

type httpWriter struct {
	http.ResponseWriter
	StatusCode int
}

const GlobalMethod = "Global"

func (w *httpWriter) WriteHeader(status int) {
	w.StatusCode = status
	w.ResponseWriter.WriteHeader(status)
}

func generateEvent(s string) string {
	buf := bytes.NewBuffer(nil)
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			continue
		}
		if ch == '/' {
			continue
		}
		buf.WriteRune(ch)
	}
	return buf.String()
}

func apiEchoHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	requestCounter.WithLabelValues(GlobalMethod, "IngressProtocol:"+formatHttpProto(r.ProtoMajor, r.ProtoMinor)).Inc()
	io.Copy(io.Discard, r.Body)
	realW := &httpWriter{ResponseWriter: w}
	w = realW
	startT := time.Now()
	var err error
	defer func() {
		event := "REQ_END:Status" + strconv.Itoa(realW.StatusCode)
		if err != nil {
			event = ";" + generateEvent(err.Error())
		}
		requestCounter.WithLabelValues(r.URL.Path, event).Inc()
		requestDuration.WithLabelValues(r.URL.Path).Observe(time.Since(startT).Seconds())
	}()
	c = httptrace.WithClientTrace(c, httpTrace)
	const urlStr = "http://127.0.0.1:8080/api/not_found"
	httpReq, err := http.NewRequestWithContext(c, http.MethodGet, urlStr, nil)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	httpResp, err := getHttpClient().Do(httpReq)
	if err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
		return
	}
	requestCounter.WithLabelValues(GlobalMethod, "EgressProtocol:"+formatHttpProto(httpResp.ProtoMajor, httpResp.ProtoMinor)).Inc()
	defer httpResp.Body.Close()
	h := w.Header()
	for k, vv := range httpResp.Header {
		h.Del(k)
		for _, v := range vv {
			h.Add(k, v)
		}
	}
	w.WriteHeader(httpResp.StatusCode)
	_, err = io.Copy(w, httpResp.Body)
}

var httpTrace = &httptrace.ClientTrace{
	ConnectDone: func(network, addr string, err error) {
		if err != nil {
			requestCounter.WithLabelValues(GlobalMethod, "ConnectErr:"+generateEvent(err.Error())).Inc()
		}
		requestCounter.WithLabelValues(GlobalMethod, "OutGressTcpConnect").Inc()
	},
	GotConn: func(info httptrace.GotConnInfo) {
		if info.Reused {
			requestCounter.WithLabelValues(GlobalMethod, "ReusedTcpConnect").Inc()
		}
	},
	PutIdleConn: func(err error) {
		if err != nil {
			requestCounter.WithLabelValues(GlobalMethod, "PutIdleConnError:"+err.Error()).Inc()
		} else {
			requestCounter.WithLabelValues(GlobalMethod, "PutIdleConnSuccess").Inc()
		}
	},
}

func formatHttpProto(major, minor int) string {
	return fmt.Sprintf("Version:%d_%d", major, minor)
}

var getHttpClient = sync.OnceValue[*http.Client](func() *http.Client {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1000,
		DialContext: func(c context.Context, network, addr string) (net.Conn, error) {
			var dialer net.Dialer
			conn, err := dialer.DialContext(c, network, addr)
			if err != nil {
				return conn, err
			}
			return &connWrapper{Conn: conn}, nil
		},
	}

	return &http.Client{Transport: transport}
})

type connWrapper struct {
	net.Conn
}

func (cw *connWrapper) Close() error {
	return cw.Conn.Close()
}
