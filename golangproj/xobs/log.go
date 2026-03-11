package xobs

import (
	"context"
	"log/slog"
)

func LogAndReport(ctx context.Context, event string, args ...any) {}

func LogRaw(ctx context.Context, skip int, lvl slog.Level, event string, args ...any) {}
