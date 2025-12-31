package httpsvr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_HandlerCall(t *testing.T) {
	h := NewHandler(func(ctx context.Context, req *string) (*string, error) {
		return req, nil
	})
	req := "hello"
	resp, err := h.call(context.Background(), &req)
	require.NoError(t, err)
	require.Equal(t, "hello", *resp.(*string))

	// call with req=nil
	resp, err = h.call(context.Background(), nil)
	require.Equal(t, "req must not nil when call handler", err.Error())
}

func Test_validateAndParseFunc(t *testing.T) {
	f := func(context.Context, *any) (*any, error) {
		return nil, nil
	}
	reqType, err := validateAndParseFunc(f)
	require.NoError(t, err)
	require.Equal(t, "*interface {}", reqType.String())
}
