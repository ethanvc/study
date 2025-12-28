package httpsvr

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestServer_Basic(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h := NewHandler(func(ctx context.Context, req *string) (*string, error) {
		return req, nil
	})
	h.ServeHTTP(w, req)
}
