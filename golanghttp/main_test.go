package golanghttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Form(t *testing.T) {
	var vals url.Values
	cb := func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		require.NoError(t, err)
		vals = req.Form
		fmt.Println("vals: ", vals)
	}
	svr := httptest.NewServer(http.HandlerFunc(cb))
	defer svr.Close()
	resp, err := svr.Client().Get(svr.URL + "/abc?a=b")
	require.NoError(t, err)
	_ = resp
	require.Equal(t, "b", vals.Get("a"))
}

type testHandler struct {
	Val int
}

func newHandler(val int) *testHandler {
	return &testHandler{Val: val}
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func findHandler(mux *http.ServeMux, p string) (*testHandler, string) {
	req := httptest.NewRequest(http.MethodGet, p, nil)
	h, pattern := mux.Handler(req)
	if pattern == "" {
		return nil, pattern
	}
	return h.(*testHandler), pattern
}

func Test_ServeMux(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/api/students/{id}", newHandler(3))
	h, pattern := findHandler(mux, "/api/students/3")
	require.NotNil(t, h)
	require.Equal(t, "/api/students/{id}", pattern)
}
