package golanggin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func Test_GinRoute(t *testing.T) {
	route := NewGinRouter()
	route.Register(http.MethodGet, "/")
	route.Register(http.MethodGet, "/api/aaa")
	route.Register(http.MethodGet, "/api/bbb/")

	var pattern string
	pattern, _ = route.Get(httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, "/", pattern)
	pattern, _ = route.Get(httptest.NewRequest(http.MethodGet, "/api/aaa", nil))
	require.Equal(t, "/api/aaa", pattern)
	pattern, _ = route.Get(httptest.NewRequest(http.MethodGet, "/api/aaa/", nil))
	require.Equal(t, "/api/aaa", pattern)
}

type GinRouter struct {
	g *gin.Engine
}

func NewGinRouter() *GinRouter {
	return &GinRouter{g: gin.New()}
}

func (router *GinRouter) Register(method string, pattern string) bool {
	router.g.Handle(method, pattern, router.serve)
	return true
}

func (router *GinRouter) serve(c *gin.Context) {
	result := c.Request.Context().Value(contextKeyRouteResult{}).(*RouteResult)
	result.Pattern = c.FullPath()
	result.Params = append(gin.Params(nil), c.Params...)
}

func (router *GinRouter) Get(req *http.Request) (string, gin.Params) {
	result := &RouteResult{}
	ctx := context.WithValue(req.Context(), contextKeyRouteResult{}, result)
	req = req.WithContext(ctx)
	u := req.URL
	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "" {
		u.Path = "/"
	}
	if u.RawPath != "" {
		u.RawPath = strings.TrimRight(u.RawPath, "/")
		if u.RawPath == "" {
			u.RawPath = "/"
		}
	}
	router.g.ServeHTTP(NewNopWriter(), req)
	return result.Pattern, result.Params
}

type RouteResult struct {
	Pattern string
	Params  gin.Params
}

type contextKeyRouteResult struct{}

type NopWriter struct {
	header http.Header
}

func NewNopWriter() *NopWriter {
	return &NopWriter{header: http.Header{}}
}

func (writer *NopWriter) Header() http.Header {
	return writer.header
}

func (writer *NopWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (writer *NopWriter) WriteHeader(statusCode int) {
}
