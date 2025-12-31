package httpsvr

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

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

func (h *Handler) Handle(ctx context.Context, info *CallInfo) (err error) {
	var req, resp any
	ctx = info.Server.getLogger().Start(ctx, info)
	defer func() {
		info.Server.getLogger().End(ctx, err, req, resp, info)
	}()
	resp, err = h.unmarshal(ctx, info)
	if err != nil {
		h.marshal(ctx, err, nil, info)
		return err
	}
	resp, err = h.handleRest(ctx, req, info)
	return nil
}

func (h *Handler) handleRest(ctx context.Context, req any, info *CallInfo) (any, error) {
	next := &Next{
		i:            0,
		handler:      h,
		interceptors: h.getInterceptors(info.Server),
	}
	return next.Next(ctx, req, info)
}

func (h *Handler) getInterceptors(s *Server) []Interceptor {
	if h.Interceptors != nil {
		return h.Interceptors
	}
	return s.Interceptors
}

func (h *Handler) unmarshal(ctx context.Context, info *CallInfo) (any, error) {
	return nil, nil
}

func (h *Handler) marshal(ctx context.Context, err error, resp any, info *CallInfo) {

}

func (h *Handler) NameOfFunc() string {
	return nameOfFunction(h.f.Interface())
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
