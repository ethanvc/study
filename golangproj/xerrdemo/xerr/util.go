package xerr

import (
	"errors"

	"google.golang.org/grpc/codes"
)

func Code(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	var realErr *Error
	if errors.As(err, &realErr) {
		return realErr.GetCode()
	}
	return codes.Unknown
}

func BlockBusinessErr(err error) error {
	realErr, ok := err.(*Error)
	if !ok {
		return err
	}
	if realErr == nil {
		return nil
	}
	switch realErr.GetCode() {
	case codes.Unknown, codes.Internal, codes.DeadlineExceeded, codes.Aborted,
		codes.Unimplemented, codes.Unavailable, codes.DataLoss:
		return err
	default:
		oldCode := realErr.GetCode()
		realErr.Code = codes.Internal
		realErr.AppendEvent("BlockedCode:" + oldCode.String())
		return realErr
	}
}
