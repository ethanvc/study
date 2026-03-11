package xobs

import (
	"context"
	"log/slog"
)

func Log(ctx context.Context, skip int, lvl slog.Level, event string, args ...any) {}
