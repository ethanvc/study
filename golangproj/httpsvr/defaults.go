package httpsvr

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
)

var DefaultLogger = &LoggerImpl{}

var DefaultSerializer = &JsonSerializer{}

type LoggerImpl struct{}

func (l *LoggerImpl) Start(ctx context.Context, info *CallInfo) context.Context {
	return ctx
}

func (l *LoggerImpl) End(ctx context.Context, err error, req, resp any, info *CallInfo) {
	lvl := slog.LevelInfo
	if err != nil {
		lvl = slog.LevelError
	}
	slog.Log(ctx, lvl, "REQ_END")
}

func (l *LoggerImpl) Log(ctx context.Context, lvl slog.Level, event string, args ...any) {
	slog.Log(ctx, lvl, event, args...)
}

type JsonSerializer struct {
}

func (j *JsonSerializer) Marshal(ctx context.Context, err error, v any, info *CallInfo) (int, io.ReadCloser, error) {
	buf, newErr := json.Marshal(v)
	if newErr != nil {
		return 0, nil, newErr
	}
	info.RespHeader.Set("content-type", "application/json")
	return 0, io.NopCloser(bytes.NewReader(buf)), nil
}
func (j *JsonSerializer) GetResponseStatusCode(ctx context.Context, err error) int {
	return 0
}
func (j *JsonSerializer) Unmarshal(ctx context.Context, v any, info *CallInfo) error {
	err := json.Unmarshal(info.RequestBody, v)
	return err
}
