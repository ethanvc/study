package xobs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ethanvc/study/golangproj/logjson"
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
	if item.NumAttrs() == 0 {
		state.buf.WriteByte('\n')
		h.writer.Write(state.buf.Bytes())
		return
	}
	state.buf.WriteByte('|')

	enc := state.enc
	enc.WriteToken(logjson.BeginObject)
	item.Attrs(func(a Attr) bool {
		enc.WriteToken(logjson.TokenString(a.Key))
		writeAttrValue(enc, a.Val)
		return true
	})
	enc.WriteToken(logjson.EndObject)
	state.buf.Write(state.attrBuf.Bytes())

	h.writer.Write(state.buf.Bytes())
}

func writeAttrValue(enc *logjson.Encoder, v Value) {
	v = v.Resolve()
	switch v.Kind() {
	case KindString:
		enc.WriteToken(logjson.TokenString(v.String()))
	case KindInt64:
		enc.WriteToken(logjson.TokenInt(v.Int64()))
	case KindUint64:
		enc.WriteToken(logjson.TokenUint(v.Uint64()))
	case KindFloat64:
		enc.WriteToken(logjson.TokenFloat(v.Float64()))
	case KindBool:
		enc.WriteToken(logjson.TokenBool(v.Bool()))
	case KindDuration:
		enc.WriteToken(logjson.TokenString(v.Duration().String()))
	case KindTime:
		enc.WriteToken(logjson.TokenString(v.Time().Format(time.RFC3339Nano)))
	case KindAny:
		enc.WriteToken(logjson.TokenString(fmt.Sprint(v.Any())))
	default:
		enc.WriteToken(logjson.Null)
	}
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
	buf     bytes.Buffer
	attrBuf bytes.Buffer
	enc     *logjson.Encoder
}

func newLogState() *logState {
	s := &logState{}
	s.enc = logjson.NewEncoderOf(&s.attrBuf)
	return s
}

func (s *logState) Reset() {
	s.buf.Reset()
	s.attrBuf.Reset()
	s.enc.Reset(&s.attrBuf, logjson.AllowDuplicateNames(true))
}
