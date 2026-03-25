package xobs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Level int

type LogItem struct {
	ctx *ObsContext
	lvl Level
}

func (l *LogItem) Str(key string, val string) *LogItem {
	if l == nil {
		return nil
	}
	return l
}

func (l *LogItem) Emit(event string) {
	if l == nil {
		return
	}
}

func (l *LogItem) EmitReport(event string) {
	if l == nil {
		return
	}
}

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
	panic("never come here")
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
	callerPos := GetCallerPosition(skip + 1)
	record := slog.NewRecord(time.Now(), lvl, event, 0)
	record.Add(slog.SourceKey, callerPos)
	record.Add(args...)
	l.h.Handle(ctx, record)
}

// newDefaultLogHandler 创建默认日志输出：优先写入 log/obs.log，失败则退回 stderr。
func newDefaultLogHandler() slog.Handler {
	if err := os.MkdirAll("log", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: create log directory: %v\n", err)
		return slog.NewJSONHandler(os.Stderr, nil)
	}
	f, err := os.OpenFile("log/obs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: open log file: %v\n", err)
		return slog.NewJSONHandler(os.Stderr, nil)
	}
	return slog.NewJSONHandler(f, nil)
}

var defaultObCtx = &ObsContext{
	span: &Span{
		name: "default",
	},
	logger: &Logger{h: newDefaultLogHandler()},
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
