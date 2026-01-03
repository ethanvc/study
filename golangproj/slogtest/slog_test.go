package slogtest

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_Slog(t *testing.T) {
	err := errors.New("test error")
	// %+v support call error.Error() function
	msg := fmt.Sprintf("%+v", err)
	require.Equal(t, "test error", msg)

	err = status.New(codes.NotFound, "xxxx").Err()
	msg = fmt.Sprintf("%+v", err)
	require.Equal(t, "rpc error: code = NotFound desc = xxxx", msg)
}
