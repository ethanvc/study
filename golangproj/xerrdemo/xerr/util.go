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
