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

func LogAccess(ctx context.Context, err error, req, resp any, args...any) {

}

func ReportAccess(ctx context.Context, err error, labels ...string) {

}

func LogRaw(ctx context.Context, skip int, lvl slog.Level, event string, args ...any) {}
