package experimental

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ethanvc/study/golangproj/xerr"
	"google.golang.org/grpc/codes"
)

func Log(ctx context.Context, lvl slog.Level, event string, args ...any) {

}

func Test_Case1(t *testing.T) {
	f := func(ctx context.Context, req any) (any, error) {
		hasFatalErr := true
		if hasFatalErr {
			return nil, xerr.New(codes.Internal, "Stage1Err")
		}
		return nil, nil
	}
	_, _ = f(context.Background(), nil)
}
