package golangapirouter

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Basic(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest("GET", "/", nil)
	proto := NewHttpProtocolSpec(req)
	executor := NewExecutor()
	state := &ExecutionState{
		ProtocolSpec: proto,
	}
	err := executor.Execute(ctx, state)
	require.NoError(t, err)
}
