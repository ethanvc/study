package golanggrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func Test_PassByHeader(t *testing.T) {
	ctx := context.Background()
	_, conn, cancel := createPair(t)
	defer cancel()

	cli := NewGreeterClient(conn)
	var header metadata.MD
	// pass meta data to server
	ctx = metadata.AppendToOutgoingContext(ctx, "X-TokenTo", "bearer token123")
	resp, err := cli.Echo(ctx, &EchoContent{
		Msg: "hello",
	}, grpc.Header(&header))
	require.NoError(t, err)
	require.Equal(t, "hello", resp.Msg)
	require.Equal(t, "bearer token123", getFirstInMetadata(header, "X-TokenTo"))
}

func getFirstInMetadata(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}
