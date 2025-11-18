package golanghttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Http2(t *testing.T) {
	resp, err := http.Get("https://baidu.com")
	require.NoError(t, err)
	defer resp.Body.Close()
	_ = resp
}
