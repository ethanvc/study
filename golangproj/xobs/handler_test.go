package xobs

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopWriteCloser struct {
	bytes.Buffer
}

func (*nopWriteCloser) Close() error { return nil }

func TestJsonHandler_Handle(t *testing.T) {
	var writer nopWriteCloser
	handler := NewJsonHandler(&writer)
	ctx := WithObsContext(context.Background(), &ObsConfig{Handler: handler})
	LogInfo(ctx, "test")
	require.Equal(t, "2026-01-01T00:00:00Z|info|xobs/handler_test.go:23|1234567890:1234567890:1234567890|test\n", writer.String())
	writer.Reset()
	LogInfo(ctx, "test", String("key", "value"))
	require.Equal(t, `2026-01-01T00:00:00Z|info|xobs/handler_test.go:26|1234567890:1234567890:1234567890|test|{"key":"value"}`+"\n", writer.String())
}

func TestGetCallerPosition(t *testing.T) {
	pos := GetCallerPosition(0)
	assert.True(t, strings.HasPrefix(pos, "xobs/handler_test.go:"))
}
