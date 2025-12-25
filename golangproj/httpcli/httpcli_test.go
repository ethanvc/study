package httpcli

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_Do(t *testing.T) {
	ctx := context.Background()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write(body)
	}))
	defer svr.Close()
	var tmpStr string
	err := Do(ctx, svr.URL, "GET", &tmpStr, nil)
	require.NoError(t, err)
	require.Equal(t, "GET", tmpStr)
}
