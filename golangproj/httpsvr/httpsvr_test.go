package httpsvr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethanvc/study/golangproj/httpcli"
	"github.com/stretchr/testify/require"
)

func TestServer_Basic(t *testing.T) {
	ctx := context.Background()
	mux := http.NewServeMux()
	httpSvr := httptest.NewServer(mux)
	defer httpSvr.Close()
	svr := &Server{}
	b := &Builder{
		Svr: svr,
		Mux: mux,
	}
	Register(b, "", func(ctx context.Context, req *string) (*string, error) {
		return req, nil
	})
	resp, err := httpcli.DoType[string](ctx, httpSvr.URL, "hello", nil)
	require.NoError(t, err)
	require.Equal(t, "hello", resp)
}
