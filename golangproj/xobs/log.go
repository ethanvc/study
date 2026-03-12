package xobs

import (
	"context"
)

func LogReportErr(ctx context.Context, event string, args ...any) {
}

func LogErr(ctx context.Context, event string, args ...any) {

}

func ReportErr(ctx context.Context, event string, labels ...KV) {

}

type ObsContext struct {
	span *Span
}

type ctxKeyObsContext struct{}

func GetObsContext(ctx context.Context) *ObsContext {
	val, _ := ctx.Value(ctxKeyObsContext{}).(*ObsContext)
	return val
}

func (oc *ObsContext) ReportLogAccess(err error, req, resp any, labels []KV, args ...any) {
}

func (oc *ObsContext) SetAttr(key string, val any) {}

type Span struct {
	name string
}

type KV struct {
	Key string
	Val string
}
