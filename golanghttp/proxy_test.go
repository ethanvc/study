package golanghttp

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/http/httpproxy"
)

func Test_Proxy(t *testing.T) {
	verifyNoProxy(t, "192.0.0.0/24", "10.56.88.88", true)
	verifyNoProxy(t, "10.0.0.1", "10.0.0.1", true)
	verifyNoProxy(t, ".xx.com", "www.xx.com", true)
	verifyNoProxy(t, "10.0.0.1/24", "10.56.88.88", true)
	verifyNoProxy(t, "*", "10.56.88.88", true)
}

func verifyNoProxy(t *testing.T, noProxy string, u string, noProxyUsed bool) {
	conf := httpproxy.Config{
		NoProxy:    noProxy,
		HTTPProxy:  "http://127.0.0.1:10000",
		HTTPSProxy: "https://127.0.0.1:10000",
	}
	f := conf.ProxyFunc()
	parsedU, err := url.Parse(u)
	require.NoError(t, err)
	realU, err := f(parsedU)
	require.NoError(t, err)
	if !noProxyUsed {
		require.Equal(t, "127.0.0.1:10000", realU.Host)
	} else {
		require.Nil(t, realU)
	}
}
