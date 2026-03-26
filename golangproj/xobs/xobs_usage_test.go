package xobs

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

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
		// add something to print in access log.
		ctx := context.Background()
		GetObsContext(ctx).SetAttr("http.method", "POST")
	}

	{
		// log and return error
		f := func(ctx context.Context) (int, error) {
			var req, resp any
			err := errors.New("internal error")
			if err != nil {
				// will decide log level internally.
				return 0, New(codes.Unknown, "CallSvrAErr").SetMsg(err.Error()).
					LogReport(ctx, "req", req, "resp", resp)
			}
			return 0, nil
		}
		_, _ = f(context.Background())
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
	{
		// case: middleware use this to print access log and do monitor report.
		// it's business's responsibility to add more content to print in log.
		// but how to add more report labels?
		ctx := context.Background()
		var req, resp any
		var reqHeader http.Header
		err := errors.New("some error")
		// implementation will add method and timecost in the log.
		GetObsContext(ctx).ReportAccessLog(err, req, resp, nil, "req_header", reqHeader)
	}
}
