package httpsvr

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	Serializer   Serializer
	Timeout      time.Duration
	Interceptors []Interceptor
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

func NewHandler[Req any, Resp any](h func(ctx context.Context, req *Req) (*Resp, error)) *Handler {
	hh := &Handler{}
	hh.realH = func(ctx context.Context, method string, req any, info *CallInfo) (any, error) {
		realReq, ok := req.(*Req)
		if !ok {
			return nil, fmt.Errorf("FatalErr;InvalidRequestType;t=%T", req)
		}
		return h(ctx, realReq)
	}
	hh.NewReq = func() any {
		return new(Req)
	}
	return hh
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
