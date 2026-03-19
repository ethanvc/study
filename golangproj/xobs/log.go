package xobs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

func LogInfo(ctx context.Context, event string, args ...any) {
	obsCtx := GetObsContext(ctx)
	obsCtx.GetLogger().LogRaw(ctx, obsCtx, 1, slog.LevelInfo, event, args...)
}

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
	h slog.Handler
}

func (l *Logger) LogRaw(ctx context.Context, obsCtx *ObsContext, skip int, lvl slog.Level, event string, args ...any) {
	if !obsCtx.Enabled(lvl) {
		return
	}
	pc, _, _, _ := runtime.Caller(skip + 1)
	record := slog.NewRecord(time.Now(), lvl, event, pc)
	record.Add(args...)
	l.h.Handle(ctx, record)
}

var defaultObCtx = &ObsContext{
	span: &Span{
		name: "default",
	},
}

var defaultLogger = &Logger{
	h: slog.NewJSONHandler(os.Stderr, nil),
}

func init() {
	if err := os.MkdirAll("log", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: create log directory: %v\n", err)
		return
	}
	f, err := os.OpenFile("log/obs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: open log file: %v\n", err)
		return
	}
	defaultLogger = &Logger{
		h: slog.NewJSONHandler(f, nil),
	}
}
