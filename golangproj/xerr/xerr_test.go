package xerr

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestError_Usage(t *testing.T) {
	// code used for condition check in program code, api should only return predefined codes, please reference
	// for details:
	err := New(codes.NotFound, "TxnNotFound")
	require.Equal(t, "TxnNotFound", err.GetEvent()[0])
	err.SetMsg("hello")
	require.Equal(t, "hello", err.GetMsg())
	err.SetMsgf("%d", 3)
	require.Equal(t, "3", err.GetMsg())
}
