package golanghttp

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/http/httpproxy"
)

func Test_Proxy(t *testing.T) {
	// if no schema, no proxy will used.

	// must use valid cidr address format
	verifyNoProxy(t, "192.168.3.0/8", "http://10.56.88.88", true)
	verifyNoProxy(t, "192.168.3.0/8", "http://192.168.3.1", false)
	verifyNoProxy(t, "192.168.3.0/16", "http://192.168.4.1", false)
	verifyNoProxy(t, "10.0.0.1", "http://10.0.0.1", false)
	verifyNoProxy(t, ".xx.com", "http://www.xx.com", false)
	verifyNoProxy(t, "*", "http://10.56.88.88", false)
}

func verifyNoProxy(t *testing.T, noProxy string, u string, useProxy bool) {
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
	if useProxy {
		require.Equal(t, "127.0.0.1:10000", realU.Host)
	} else {
		require.Nil(t, realU)
	}
}
