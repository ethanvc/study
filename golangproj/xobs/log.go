package xobs

import (
	"context"
	"log/slog"
)

func LogReportErr(ctx context.Context, event string, args ...any) {
	LogErr(ctx, event, args...)
	ReportErr(ctx, event)
}

func LogErr(ctx context.Context, event string, args ...any) {

}

func ReportErr(ctx context.Context, event string, labels ...string) {

}

func LogAccess(ctx context.Context, err error, req, resp any, args ...any) {

}

func ReportAccess(ctx context.Context, err error, labels ...string) {

}

func LogRaw(ctx context.Context, skip int, lvl slog.Level, event string, args ...any) {}

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

type Span struct {
	name string
}

type KV struct {
	Key string
	Val string
}
