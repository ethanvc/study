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
	require.Equal(t, "test|INFO|?:0|0:0:0|test|{}", writer.String())
}

func TestGetCallerPosition(t *testing.T) {
	pos := GetCallerPosition(0)
	assert.True(t, strings.HasPrefix(pos, "xobs/handler_test.go:"))
}
