package xerr

import (
	"bytes"

	"google.golang.org/grpc/codes"
)

type Error struct {
	code  codes.Code
	event []string
	msg   string
	err   error
}

func New(code codes.Code, event string) *Error {
	err := &Error{
		code: code,
	}
	if event != "" {
		err.event = []string{event}
	}
	return err
}

const delimiter = ';'

func (e *Error) GerReportEvent() string {
	if e == nil {
		return codes.OK.String()
	}
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.code.String())
	for _, event := range e.event {
		buf.WriteByte(delimiter)
		buf.WriteString(event)
	}
	return buf.String()
}

func (e *Error) SetErr(err error) *Error {
	e.err = err
	return e
}

func (e *Error) Error() string {
	return ""
}
