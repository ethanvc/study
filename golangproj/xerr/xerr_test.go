package xerr

import (
	"testing"

	"google.golang.org/grpc/codes"
)

func TestError_Usage(t *testing.T) {
	// code used for condition check in program code, api should only return predefined codes, please reference
	// for details:
	err := New(codes.NotFound, "TxnNotFound")
	_ = err
}
