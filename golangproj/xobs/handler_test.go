package xobs

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type nopWriteCloser struct {
	bytes.Buffer
}

func (nopWriteCloser) Close() error { return nil }

func TestJsonHandler_Handle(t *testing.T) {
	var buf nopWriteCloser
	h := NewJsonHandler(&buf)

	now := time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)
	ctx := WithSpanContext(context.Background(), &SpanConfig{
		Name:         "TestSpan",
		TraceId:      "trace-001",
		SpanId:       "span-002",
		ParentSpanId: "span-001",
	})

	item := LogItem{
		Msg:      "order created",
		Time:     now,
		Level:    LevelInfo,
		Position: "xobs/handler.go:24",
		ObsCtx:   GetObsContext(ctx),
	}
	item.AddAttrs(
		String("userId", "u-123"),
		Int("count", 42),
		Bool("ok", true),
		Float64("score", 3.14),
		Duration("elapsed", 150*time.Millisecond),
	)

	h.Handle(ctx, item)

	output := buf.String()
	t.Log("output:", output)

	parts := strings.SplitN(output, "|", 6)
	assert.Len(t, parts, 6)

	assert.Equal(t, now.Format(time.RFC3339Nano), parts[0])
	assert.Equal(t, "Info", parts[1])
	assert.Equal(t, "xobs/handler.go:24", parts[2])
	assert.Equal(t, "trace-001:span-002:span-001", parts[3])
	assert.Equal(t, "order created", parts[4])

	jsonPart := parts[5]
	assert.Contains(t, jsonPart, `"userId":"u-123"`)
	assert.Contains(t, jsonPart, `"count":42`)
	assert.Contains(t, jsonPart, `"ok":true`)
	assert.Contains(t, jsonPart, `"score":3.14`)
	assert.Contains(t, jsonPart, `"elapsed":"150ms"`)
	assert.True(t, strings.HasSuffix(jsonPart, "\n"))
}

func TestJsonHandler_Handle_NoAttrs(t *testing.T) {
	var buf nopWriteCloser
	h := NewJsonHandler(&buf)

	item := LogItem{
		Msg:      "simple",
		Time:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Level:    LevelErr,
		Position: "test:1",
		ObsCtx:   defaultObCtx,
	}

	h.Handle(context.Background(), item)

	output := buf.String()
	t.Log("output:", output)

	assert.Contains(t, output, "Err")
	assert.Contains(t, output, "simple")
	assert.Contains(t, output, "{}\n")
}

func TestJsonHandler_Flush(t *testing.T) {
	pr, pw := io.Pipe()
	h := NewJsonHandler(pw)
	h.Flush()

	_, err := pr.Read(make([]byte, 1))
	assert.Equal(t, io.EOF, err)
}
