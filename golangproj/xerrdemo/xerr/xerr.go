package xerr

import (
	"bytes"

	"google.golang.org/grpc/codes"
)

type Error struct {
	// DO NOT ACCESS DIRECTLY. here use public field only for marshal/unmarshal
	Code  codes.Code
	Event []string
	Msg   string
	Err   error
}

func New(code codes.Code, event string) *Error {
	err := &Error{
		Code: code,
	}
	if event != "" {
		err.Event = []string{event}
	}
	return err
}

const delimiter = ';'

func (e *Error) GerReportEvent() string {
	if e == nil {
		return codes.OK.String()
	}
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Code.String())
	for _, event := range e.Event {
		buf.WriteByte(delimiter)
		buf.WriteString(event)
	}
	return buf.String()
}

func (e *Error) SetErr(err error) *Error {
	e.Err = err
	return e
}

func (e *Error) Error() string {
	return ""
}
