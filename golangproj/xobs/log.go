package xobs

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

func LogInfo(ctx context.Context, event string, args ...any) {
	obsCtx := GetObsContext(ctx)
	obsCtx.LogRaw(ctx, obsCtx, 1, LevelInfo, event, args...)
}

func LogReportErr(ctx context.Context, event string, args ...any) {
}

func LogErr(ctx context.Context, event string, args ...any) {
	obsCtx := GetObsContext(ctx)
	obsCtx.LogRaw(ctx, obsCtx, 1, LevelErr, event, args...)
}

func ReportErr(ctx context.Context, event string, labels ...KV) {

}

type GetLogLevelType func(err error) Level

type ObsContext struct {
	parent      *ObsContext
	span        *Span
	handler     Handler
	lvl         Level
	getLogLevel GetLogLevelType
}

type ctxKeyObsContext struct{}

type SpanConfig struct {
	Name         string
	TraceId      string
	SpanId       string
	ParentSpanId string
	GetLogLevel  GetLogLevelType
}

type ObsConfig struct {
	Handler     Handler
	GetLogLevel GetLogLevelType
	Level       Level
}

func WithSpanContext(ctx context.Context, config *SpanConfig) context.Context {
	span := &Span{}
	span.init(ctx, config)
	ctx, obsCtx := withObsContext(ctx)
	obsCtx.span = span
	return ctx
}

func WithObsContext(ctx context.Context, config *ObsConfig) context.Context {
	obsCtx := &ObsContext{
		getLogLevel: config.GetLogLevel,
		lvl:         config.Level,
		handler:     config.Handler,
	}
	return context.WithValue(ctx, ctxKeyObsContext{}, obsCtx)
}

func withObsContext(ctx context.Context) (context.Context, *ObsContext) {
	obsCtx := &ObsContext{}
	return context.WithValue(ctx, ctxKeyObsContext{}, obsCtx), obsCtx
}

func GetObsContext(ctx context.Context) *ObsContext {
	val, _ := ctx.Value(ctxKeyObsContext{}).(*ObsContext)
	if val == nil {
		return defaultObCtx
	}
	return val
}

func GetRootSpan(ctx context.Context) *Span {
	obsCtx := GetObsContext(ctx)
	return obsCtx.GetRootSpan()
}

func (oc *ObsContext) GetRootSpan() *Span {
	var span *Span
	for oc != nil {
		if oc.span != nil {
			span = oc.span
		}
		oc = oc.parent
	}
	if span != nil {
		return span
	}
	return defaultObCtx.span
}

func (oc *ObsContext) GetSpan() *Span {
	for oc != nil {
		if oc.span != nil {
			return oc.span
		}
		oc = oc.parent
	}
	return nil
}

func (oc *ObsContext) LogReportAccessLog(err error, req, resp any, labels []KV, args ...any) {
}

func (oc *ObsContext) SetAttr(key string, val any) {}

func (oc *ObsContext) GetHandler() Handler {
	for oc != nil {
		if oc.handler != nil {
			return oc.handler
		}
		oc = oc.parent
	}
	panic("never come here")
}

func (oc *ObsContext) GetLevel() Level {
	for oc != nil {
		if oc.lvl != LevelNotSet {
			return oc.lvl
		}
		oc = oc.parent
	}
	return LevelInfo
}

func (oc *ObsContext) Enabled(lvl Level) bool {
	return lvl >= oc.GetLevel()
}

func (oc *ObsContext) LogRaw(ctx context.Context, obsCtx *ObsContext, skip int, lvl Level, event string, args ...any) {
	if !oc.Enabled(lvl) {
		return
	}
	item := LogItem{
		Msg:      event,
		Time:     time.Now(),
		Level:    lvl,
		Position: GetCallerPosition(skip),
		ObsCtx:   obsCtx,
	}
	item.Add(args...)
	obsCtx.GetHandler().Handle(ctx, item)
}

type KV struct {
	Key string
	Val string
}

type Handler interface {
	Handle(ctx context.Context, item LogItem)
	Flush()
}

var defaultObCtx = &ObsContext{
	span: &Span{
		name: "default",
	},
}

// lastTwoPathParts 返回路径的最后两段（如 pkg/foo.go）；不足两段则只返回最后一段。
func lastTwoPathParts(file string) string {
	file = filepath.Clean(file)
	base := filepath.Base(file)
	dir := filepath.Dir(file)
	if dir == "." {
		return base
	}
	parent := filepath.Base(dir)
	if parent == "." || parent == string(filepath.Separator) {
		return base
	}
	return filepath.Join(parent, base)
}

func GetCallerPosition(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "?:0"
	}
	return fmt.Sprintf("%s:%d", lastTwoPathParts(file), line)
}
