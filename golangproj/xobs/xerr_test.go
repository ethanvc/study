package xobs

import (
	"context"
	"encoding/json"
	"errors"
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

func Test_Case(t *testing.T) {
	{
		// case: return error and let middleware print it in log and do monitor report.
		f := func(ctx context.Context, req string) (int, error) {
			type Abc struct {
				A string
			}
			var objReq Abc
			err := json.Unmarshal([]byte(req), &objReq)
			if err != nil {
				// return error to let middleware process
				return 0, New(codes.InvalidArgument, "ArgumentNotValidJson").SetMsg(err.Error())
			}
			return 0, nil
		}
		_, err := f(context.Background(), "")
		require.Equal(t, codes.InvalidArgument, Code(err))
	}

	{
		// case: log and report in-place, but the error can downgrade.
		f := func(ctx context.Context) error {
			err := errors.New("some error")
			if err != nil {
				LogReportErr(ctx, "keyOperationFailedButDowngraded", "err", err)
			}
			return nil
		}
		_ = f(context.Background())
	}
}
