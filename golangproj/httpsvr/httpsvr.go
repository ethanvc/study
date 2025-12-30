package httpsvr

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"
)

type Builder struct {
	Svr *Server
	Mux *http.ServeMux
}

func (b *Builder) Register(pattern string, h any) {

}

type Server struct {
	Serializer   Serializer
	Logger       Logger
	Timeout      time.Duration
	Interceptors []Interceptor
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) NewHttpHandler(h *Handler) http.Handler {
	hh := func(w http.ResponseWriter, r *http.Request) {
		s.serverHttp(h, w, r)
	}
	return http.HandlerFunc(hh)
}

func (s *Server) serverHttp(h *Handler, w http.ResponseWriter, r *http.Request) {

}

type Interceptor func(ctx context.Context, pattern string, req any, info *CallInfo, next Next) (any, error)

type Next struct {
	i            int
	interceptors []Interceptor
	handler      *Handler
}

func (n Next) Next(ctx context.Context, pattern string, req any, info *CallInfo) (any, error) {
	if n.i >= len(n.interceptors) {
		return n.handler.realH(ctx, pattern, req, info)
	}
	newN := n
	newN.i++
	return n.interceptors[n.i](ctx, pattern, req, info, newN)
}

type Handler struct {
	Serializer   Serializer
	Timeout      time.Duration
	Interceptors []Interceptor
	realH        func(ctx context.Context, method string, req any, info *CallInfo) (any, error)
	NewReq       func() any
}

func NewHandler(f any) *Handler {
	hh := &Handler{}
	return hh
}

func (h *Handler) init(f any) {

}

func validateAndParseFunc(f any) (reqType, respType reflect.Type, err error) {
	funcType := reflect.TypeOf(f)
	if funcType.Kind() != reflect.Func {
		return nil, nil, fmt.Errorf("the input must be a function type: %v", funcType.Kind())
	}

	if funcType.NumIn() != 2 {
		return nil, nil, fmt.Errorf("the function must have exactly two parameter")
	}

	firstParam := funcType.In(0)
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

	if firstParam != contextType {
		return nil, nil, fmt.Errorf("the first parameter must be context.Context")
	}

	secondParam := funcType.In(1)
	if secondParam.Kind() != reflect.Ptr {
		return nil, nil, fmt.Errorf("the second parameter must be a pointer")
	}

	if funcType.NumOut() != 2 {
		return nil, nil, fmt.Errorf("function must return exact two params")
	}

	firstReturn := funcType.Out(0)
	if firstReturn.Kind() != reflect.Ptr {
		return nil, nil, fmt.Errorf("the first return param must be a pointer")
	}

	secondReturn := funcType.Out(1)
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if secondReturn != errorType {
		return nil, nil, fmt.Errorf("the second return param must be error type")
	}

	return secondParam.Elem(), firstReturn.Elem(), nil
}

type CallInfo struct {
	Request *http.Request
	Writer  http.ResponseWriter
	Server  *Server
}

type Serializer interface {
	Marshal(ctx context.Context, v any, info *CallInfo) (string, []byte, error)
	Unmarshal(ctx context.Context, v any, info *CallInfo) error
}

type Logger interface{}
