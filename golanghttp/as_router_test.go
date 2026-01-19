package golanghttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	pattern := ""
	r := NewRouter()
	r.Register("/api/bcd")
	r.Register("POST www.xx.com/api/bcd")
	r.Get(http.MethodGet, "/api/bcd")
	pattern = r.Get(http.MethodGet, "/api/bcd")
	require.Equal(t, "/api/bcd", pattern)
	pattern = r.Get(http.MethodPost, "http://www.xx.com/api/bcd")
	require.Equal(t, "/api/bcd", pattern)
}

type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	r := &Router{
		mux: http.NewServeMux(),
	}
	return r
}

func (r *Router) Register(pattern string) {
	r.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {})
}

func (r *Router) Get(method, urlStr string) (pattern string) {
	// 使用 mux.Handler 方法获取 pattern（这是导出的方法，不需要反射）
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return ""
	}
	// Handler 方法返回 (Handler, pattern string)
	_, pattern = r.mux.Handler(req)
	return pattern
}
