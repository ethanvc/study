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
