package xobs

import (
	"context"
	"log/slog"
)

func LogReportErr(ctx context.Context, event string, args ...any) {
}

func LogErr(ctx context.Context, event string, args ...any) {
	obsCtx := GetObsContext(ctx)
	obsCtx.GetLogger().LogRaw(ctx, obsCtx, 1, slog.LevelError, event, args...)
}

func ReportErr(ctx context.Context, event string, labels ...KV) {

}

type ObsContext struct {
	parent *ObsContext
	span   *Span
	logger *Logger
	lvl    *slog.Level
}

type ctxKeyObsContext struct{}

func GetObsContext(ctx context.Context) *ObsContext {
	val, _ := ctx.Value(ctxKeyObsContext{}).(*ObsContext)
	if val == nil {
		return defaultObCtx
	}
	return val
}

func (oc *ObsContext) ReportLogAccess(err error, req, resp any, labels []KV, args ...any) {
}

func (oc *ObsContext) SetAttr(key string, val any) {}

func (oc *ObsContext) GetLogger() *Logger {
	for oc != nil {
		if oc.logger != nil {
			return oc.logger
		}
		oc = oc.parent
	}
	return defaultLogger
}

func (oc *ObsContext) GetLvl() slog.Level {
	for oc != nil {
		if oc.lvl != nil {
			return *oc.lvl
		}
		oc = oc.parent
	}
	return slog.LevelInfo
}

func (oc *ObsContext) Enabled(lvl slog.Level) bool {
	return lvl >= oc.GetLvl()
}

type Span struct {
	name string
}

type KV struct {
	Key string
	Val string
}

type Logger struct {
}

func (l *Logger) LogRaw(ctx context.Context, obsCtx *ObsContext, skip int, lvl slog.Level, event string, args ...any) {
	if !obsCtx.Enabled(lvl) {
		return
	}
}

var defaultObCtx = &ObsContext{
	span: &Span{
		name: "default",
	},
}

var defaultLogger = &Logger{}
