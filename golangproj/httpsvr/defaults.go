package httpsvr

import "context"

var DefaultLogger Logger

type LoggerImpl struct{}

func (l *LoggerImpl) Start(ctx context.Context, info *CallInfo) context.Context {
	return ctx
}

func (l *LoggerImpl) getLogger() Logger {}
