package cutgo

import (
	"context"
	"errors"
)

type HandlerFunc func(ctx context.Context, req any, resp any) error

type Handler[CallInfo any] struct {
	HandlerFunc  HandlerFunc
	newReq       func() any
	newResp      func() any
	interceptors InterceptorChan[CallInfo]
}

type Interceptor[CallInfo any] func(ctx context.Context, method string, req, resp any,
	info *CallInfo, next Next[CallInfo]) error

type InterceptorChan[CallInfo any] []Interceptor[CallInfo]

type Next[CallInfo any] func(ctx context.Context, method string, req, resp any, info *CallInfo) error

type HandlerInfo[CallInfo any] struct {
	handler   func(ctx context.Context, req, resp any) error
	newReq    func() any
	newResp   func() any
	unmarshal func(ctx context.Context, info *CallInfo) (any, error)
	marshal   func(ctx context.Context, resp any, info *CallInfo) ([]byte, error)
}

func NewHandlerInfo[CallInfo any, Req, Resp any](f func(ctx context.Context, req *Req, resp *Resp) error) *HandlerInfo[CallInfo] {
	h := &HandlerInfo[CallInfo]{
		handler: func(ctx context.Context, req, resp any) error {
			realReq, ok := req.(*Req)
			if !ok {
				return errors.New("fatal error: invalid request type")
			}
			realResp, ok := resp.(*Resp)
			if !ok {
				return errors.New("fatal error: invalid response type")
			}
			return f(ctx, realReq, realResp)
		},
		newReq: func() any {
			return new(Req)
		},
		newResp: func() any {
			return new(Resp)
		},
	}
	return h
}

func (h *HandlerInfo[CallInfo]) NewReq() any {
	return h.newReq()
}

func (h *HandlerInfo[CallInfo]) NewResp() any {
	return h.newResp()
}

func (h *HandlerInfo[CallInfo]) Call(ctx context.Context, req, resp any) error {
	return h.handler(ctx, req, resp)
}
