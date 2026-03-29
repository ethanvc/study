package xobs

import (
	"context"
	"log/slog"
	"testing"

	"google.golang.org/grpc/codes"
)

func GetUserAccount(ctx context.Context, userId int64) (*UserAccount, error) {
	return nil, nil
}

func GetVoucher(ctx context.Context, req *GetVoucherReq) (*GetVoucherResp, error) {
	return nil, nil
}

type GetVoucherReq struct {
	UserId int64 `json:"user_id"`
	Amount int64 `json:"amount"`
}

type GetVoucherResp struct {
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
			Amount     int64  `json:"amount"`
			UserId     int64  `json:"user_id"`
			BusinessId int64  `json:"business_id"`
			MchOrderId string `json:"mch_order_id"`
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
				return nil, New(codes.Unknown, "CallGetUserAccountErr").SetMsg(err.Error()).
					LogReport(ctx, "req", req)
			}
			_ = account
			voucher, err := GetVoucher(ctx, &GetVoucherReq{
				UserId: req.UserId,
				Amount: req.Amount,
			})
			if err != nil {
				LogReportErr(ctx, "GetVoucherErr", "err", err)
			}
			_ = voucher
			return &CreateOrderResp{}, nil
		}
		ctx := WithSpanContext(ctx, &SpanConfig{Name: "YourApiName"})
		req := &CreateOrderReq{}
		resp, err := f(ctx, req)
		// in real case, you should get report kvs after verified.
		GetObsContext(ctx).LogReportAccessLog(err, req, resp, []KV{{Key: "business_id", Val: "333"}})
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
