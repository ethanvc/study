package httpcli

import (
	"context"
	"net/http/httptrace"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Proxy(t *testing.T) {
	// we can use this method to detect proxy usage
	os.Setenv("http_proxy", "https://www.abc.com")
	ctx := context.Background()
	target := ""
	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			target = hostPort
		},
	}
	ctx = httptrace.WithClientTrace(ctx, trace)
	err := Do(ctx, "http://www.xx.com", "hello", nil, nil)
	_ = err
	require.Equal(t, "www.abc.com:443", target)

}
