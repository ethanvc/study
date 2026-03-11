package xobs

import (
	"bytes"
	"fmt"

	"google.golang.org/grpc/codes"
)

type Payload struct {
	Key string `json:"key"`
	Val any    `json:"val"`
}

type Error struct {
	// DO NOT ACCESS DIRECTLY. here use public field only for marshal/unmarshal
	Code    codes.Code `json:"code"`
	Event   []string   `json:"event,omitempty"`
	Msg     string     `json:"msg,omitempty"`
	Details []Payload  `json:"details,omitempty"`
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

func (e *Error) GetCode() codes.Code {
	if e == nil {
		return codes.OK
	}
	return e.Code
}

func (e *Error) GetEvent() []string {
	if e == nil {
		return []string{}
	}
	return e.Event
}

func (e *Error) GetMsg() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func (e *Error) GetDetails() []Payload {
	if e == nil {
		return nil
	}
	return e.Details
}

func (e *Error) SetMsg(msg string) *Error {
	e.Msg = msg
	return e
}

func (e *Error) SetMsgf(format string, args ...any) *Error {
	e.Msg = fmt.Sprintf(format, args...)
	return e
}

func (e *Error) AppendEvent(event string) *Error {
	const maxAllowedEvent = 100
	if len(e.Event) > maxAllowedEvent {
		return e
	}
	e.Event = append(e.Event, event)
	return e
}

func (e *Error) AppendKvEvent(k string, v any) *Error {
	buf := fmt.Sprintf("%s:%v", k, v)
	return e.AppendEvent(buf)
}

const delimiter = ';'

func (e *Error) GetReportEvent() string {
	if e.GetCode() == codes.OK {
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

func (e *Error) Error() string {
	if e.GetCode() == codes.OK {
		return codes.OK.String()
	}
	return e.GetReportEvent() + ";" + e.Msg
}

func (e *Error) ReportAndLog(args ...any) *Error {
	return e
}

var errorLog = Log
