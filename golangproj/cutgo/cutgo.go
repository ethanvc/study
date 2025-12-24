package cutgo

import (
	"context"
	"errors"
)

type HandlerFunc func(ctx context.Context, req any, resp any) error

type Handler[CallInfo any] struct {
	HandlerFunc  func(ctx context.Context, req any, resp any, info *CallInfo) error
	interceptors InterceptorChain[CallInfo]
}

func NewHandler[Req, Resp, CallInfo any](hh func(ctx context.Context, req Req, resp Resp, info *CallInfo) error) *Handler[CallInfo] {
	h := &Handler[CallInfo]{}
	h.HandlerFunc = func(ctx context.Context, req any, resp any, info *CallInfo) error {
		realReq, ok := req.(Req)
		if !ok {
			return errors.New("invalid request type")
		}
		realResp, ok := resp.(Resp)
		if !ok {
			return errors.New("invalid response type")
		}
		return hh(ctx, realReq, realResp, info)
	}
	return h
}

func (h *Handler[CallInfo]) Call(ctx context.Context, method string, req any, resp any, info *CallInfo) error {
	n := Next[CallInfo]{
		idx:     0,
		handler: h,
	}
	return n.Call(ctx, method, req, resp, info)
}

type Interceptor[CallInfo any] func(ctx context.Context, method string, req, resp any,
	info *CallInfo, next Next[CallInfo]) error

type InterceptorChain[CallInfo any] []Interceptor[CallInfo]

type Next[CallInfo any] struct {
	idx     int
	handler *Handler[CallInfo]
}

func (n Next[CallInfo]) Call(ctx context.Context, method string, req any, resp any, info *CallInfo) error {
	if n.idx >= len(n.handler.interceptors) {
		return n.handler.HandlerFunc(ctx, req, resp, info)
	}
	interceptor := n.handler.interceptors[n.idx]
	newN := n
	newN.idx++
	return interceptor(ctx, method, req, resp, info, newN)
}
