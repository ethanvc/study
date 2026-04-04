package xobs

import (
	"bytes"
	"context"
	"testing"
)

type nopWriteCloser struct {
	bytes.Buffer
}

func (*nopWriteCloser) Close() error { return nil }

func TestJsonHandler_Handle(t *testing.T) {
	var writer nopWriteCloser
	handler := NewJsonHandler(&writer)
	ctx := WithObsContext(context.Background(), &ObsConfig{Ha})
}
