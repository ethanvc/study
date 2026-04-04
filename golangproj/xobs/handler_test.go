package xobs

import (
	"bytes"
	"context"
	"testing"

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
