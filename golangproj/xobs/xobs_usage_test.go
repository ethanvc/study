package xobs

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
)

func GetUserAccount(ctx context.Context, userId int64) (*UserAccount, error) {
	return nil, nil
}

type UserAccount struct {
	UserId  int64  `json:"user_id"`
	Account string `json:"account"`
	Balance int64  `json:"balance"`
}

func Test_Case(t *testing.T) {
	ctx := context.Background()
	{
		// case: normal business case
		type CreateOrderReq struct {
			Amount     int64 `json:"amount"`
			UserId     int64 `json:"user_id"`
			BusinessId int64 `json:"business_id"`
		}
		type CreateOrderResp struct{}
		f := func(ctx context.Context, req *CreateOrderReq) (*CreateOrderResp, error) {
			if req.Amount <= 0 {
				return nil, New(codes.InvalidArgument, "AmountNotValid").SetMsg("amount must be greater than 0")
			}
			// set something to print in access log and as report label.
			GetRootSpan(ctx).SetAttr("business_id", req.BusinessId)
			// access downstream and can not downgrade
			account, err := GetUserAccount(ctx, req.UserId)
			if err != nil {
				return nil, New(codes.Unknown, "CallSvrAErr").SetMsg(err.Error()).
					LogReport(ctx, "req", req, "resp", resp)
			}
			_ = account
			return &CreateOrderResp{}, nil
		}
		ctx := WithSpanContext(ctx, &SpanConfig{Name: "YourApiName"})
		req := &CreateOrderReq{}
		resp, err := f(ctx, req)
		// in real case, you should get report kvs after verified.
		GetObsContext(ctx).LogReportAccessLog(err, req, resp, []KV{{Key: "business_id", Val: "333"}})
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
		_ = f(ctx)
	}
	{
		// case: middleware use this to print access log and do monitor report.
		// it's business's responsibility to add more content to print in log.
		// but how to add more report labels?
		var req, resp any
		var reqHeader http.Header
		err := errors.New("some error")
		// implementation will add method and timecost in the log.
		GetObsContext(ctx).LogReportAccessLog(err, req, resp, nil, "req_header", reqHeader)
	}
	{
		// report with special labels
		var req, resp any
		var reqHeader http.Header
		err := errors.New("some error")
		labels := []KV{
			{Key: "user_id", Val: "333"},
		}
		GetObsContext(context.Background()).LogReportAccessLog(err, req, resp, labels, "req_header", reqHeader)
	}
	{
		// offer a function to get the log error. like redis, we want make not found as debug level, so it won't be printed in log.
		getLvl := func(err *Error) slog.Level {
			switch err.GetCode() {
			case codes.OK, codes.NotFound, codes.AlreadyExists:
				return slog.LevelDebug
			default:
				return slog.LevelError
			}
		}
		_ = getLvl
	}
}
