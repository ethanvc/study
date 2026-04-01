package xobs

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"
)

type JsonHandler struct {
	writer io.WriteCloser
}

func NewJsonHandler(writer io.WriteCloser) *JsonHandler {
	return &JsonHandler{
		writer: writer,
	}
}

func (h *JsonHandler) Handle(ctx context.Context, item LogItem) {
	state := sLogStatePool.Get().(*logState)
	defer sLogStatePool.Put(state)
	state.Reset()
	state.buf.WriteString(item.Time.Format(time.RFC3339Nano))
	state.buf.WriteByte('|')
	state.buf.WriteString(item.Level.String())
	state.buf.WriteByte('|')
	state.buf.WriteString(item.Position)
	state.buf.WriteByte('|')
	span := item.ObsCtx.GetSpan()
	state.buf.WriteString(span.GetTraceId())
	state.buf.WriteByte(':')
	state.buf.WriteString(span.GetSpanId())
	state.buf.WriteByte(':')
	state.buf.WriteString(span.GetParentSpanId())
	state.buf.WriteByte('|')
	state.buf.WriteString(item.Msg)
	state.buf.WriteByte('|')
}

func (h *JsonHandler) Flush() {
	h.writer.Close()
}

var sLogStatePool = sync.Pool{
	New: func() any {
		return newLogState()
	},
}

type logState struct {
	buf bytes.Buffer
}

func newLogState() *logState {
	return &logState{}
}

func (s *logState) Reset() {
	s.buf.Reset()
}
