package httpcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Serializer interface {
	Marshal(ctx context.Context, v any, opts *Options) (string, io.Reader, error)
	Unmarshal(ctx context.Context, httpResp *http.Response, resp any, opts *Options) error
}

type AutoSerializer struct {
}

func (s *AutoSerializer) Marshal(ctx context.Context, v any, opts *Options) (string, io.Reader, error) {
	switch realV := v.(type) {
	case string:
		return "", strings.NewReader(realV), nil
	case []byte:
		return "", bytes.NewReader(realV), nil
	default:
		buf, err := json.Marshal(v)
		if err != nil {
			return "", nil, err
		}
		return "application/json; charset=utf-8", bytes.NewReader(buf), nil
	}
}

func (s *AutoSerializer) Unmarshal(ctx context.Context, httpResp *http.Response, resp any, opts *Options) error {
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}
	opts.RespBody = body
	switch realV := resp.(type) {
	case *string:
		*realV = string(body)
	case *[]byte:
		*realV = body
	default:
		err = json.Unmarshal(body, resp)
		if err != nil {
			return fmt.Errorf("unmarshal error: %s. body is %s", err.Error(), string(body))
		}
	}
	return nil
}
