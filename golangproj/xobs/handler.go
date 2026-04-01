package xobs

import (
	"context"
	"io"
)

type JsonHandler struct {
	writer io.WriteCloser
}

func NewJsonHandler(writer io.WriteCloser) *JsonHandler {
	return &JsonHandler{
		writer: writer,
	}
}

func (h *JsonHandler) Handle(ctx context.Context, item LogItem) {

}

func (h *JsonHandler) Flush() {
	h.writer.Close()
}
