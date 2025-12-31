package httpsvr

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

type Server struct {
	Serializer   Serializer
	Logger       Logger
	Timeout      time.Duration
	Interceptors []Interceptor
}

func (s *Server) HandleFunc(pattern string, handler *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.ServeHTTP(pattern, handler, w, r)
	}
}

func (s *Server) ServeHTTP(pattern string, h *Handler, w http.ResponseWriter, r *http.Request) {

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
		return n.handler.call(ctx, req)
	}
	newN := n
	newN.i++
	return n.interceptors[n.i](ctx, pattern, req, info, newN)
}

type Handler struct {
	Serializer   Serializer
	Timeout      time.Duration
	Interceptors []Interceptor
	reqType      reflect.Type
	f            reflect.Value
}

func NewHandler(f any) *Handler {
	hh := &Handler{}
	hh.init(f)
	return hh
}

func (h *Handler) init(f any) {
	if f == nil {
		panic("must not zero for handler func")
	}
	reqType, err := validateAndParseFunc(f)
	if err != nil {
		panic(err)
	}
	h.reqType = reqType
	h.f = reflect.ValueOf(f)
}

func (h *Handler) call(ctx context.Context, req any) (any, error) {
	if req == nil {
		return nil, fmt.Errorf("req must not nil when call handler")
	}
	if reflect.TypeOf(req) != h.reqType {
		return nil, fmt.Errorf("invalid req type, expect %v, got %v", h.reqType, reflect.TypeOf(req))
	}
	result := h.f.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)})
	resp := result[0].Interface()

	var err error
	reflectErr := result[1].Interface()
	if reflectErr != nil {
		err = reflectErr.(error)
	}
	return resp, err
}

func validateAndParseFunc(f any) (reqType reflect.Type, err error) {
	funcType := reflect.TypeOf(f)
	if funcType.Kind() != reflect.Func {
		return nil, fmt.Errorf("the input must be a function type: %T", f)
	}

	if funcType.NumIn() != 2 {
		return nil, fmt.Errorf("the function must have exactly two parameter")
	}

	firstParam := funcType.In(0)
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

	if firstParam != contextType {
		return nil, fmt.Errorf("the first parameter must be context.Context")
	}

	secondParam := funcType.In(1)
	if secondParam.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("the second parameter must be a pointer")
	}

	if funcType.NumOut() != 2 {
		return nil, fmt.Errorf("function must return exact two params")
	}

	firstReturn := funcType.Out(0)
	if firstReturn.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("the first return param must be a pointer")
	}

	secondReturn := funcType.Out(1)
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if secondReturn != errorType {
		return nil, fmt.Errorf("the second return param must be error type")
	}

	return secondParam, nil
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

func nameOfFunction(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
