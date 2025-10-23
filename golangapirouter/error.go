package golangapirouter

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
)

var ErrRuleNotFound = errors.New("RuleNotFound")
var ErrNoValidUnitDecided = errors.New("NoValidUnitDecided")

type Error struct {
	code  codes.Code
	event string
	msg   string
}

func NewErr(code codes.Code, event string, args ...any) *Error {
	return &Error{
		code:  code,
		event: fmt.Sprintf(event, args...),
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, event: %s, msg: %s", e.code, e.event, e.msg)
}

func (e *Error) String() string {
	return e.Error()
}
