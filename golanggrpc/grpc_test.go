package golanggrpc

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/ringhash"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	_ "google.golang.org/grpc/xds"
)

type server struct {
	UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	// pair with client side: metadata.AppendToOutgoingContext(ctx, "authorization", "bearer token123")
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		log.Printf("Metadata: %v", md)
	}
	trailer := metadata.Pairs(
		"processing-time", "100ms",
		"end-time", time.Now().Format(time.RFC3339Nano),
	)

	header := metadata.Pairs("server-id", "srv-001", "content-type", "application/protobuf")
	err := grpc.SetHeader(ctx, header)
	if err != nil {
		return nil, err
	}
	err = grpc.SetTrailer(ctx, trailer)
	if err != nil {
		return nil, err
	}
	return &HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *server) SayHelloReturnNil(_ context.Context, in *HelloRequest) (*HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return nil, nil
}

func (s *server) Echo(ctx context.Context, in *EchoContent) (*EchoContent, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		err := grpc.SetHeader(ctx, md)
		if err != nil {
			return nil, err
		}
	}
	return in, nil
}

func createPair(t *testing.T, opts ...grpc.DialOption) (*grpc.Server, *grpc.ClientConn, func()) {
	s := grpc.NewServer()
	RegisterGreeterServer(s, &server{})
	ln, err := net.Listen("tcp", ":22222")
	require.NoError(t, err)
	go func() {
		err := s.Serve(ln)
		if err != nil {
			panic(err)
		}
	}()
	realOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	realOpts = append(realOpts, opts...)
	conn, err := grpc.NewClient("localhost:22222", realOpts...)
	require.NoError(t, err)
	return s, conn, func() {
		s.Stop()
		conn.Close()
	}
}

func Test_IdleTimer(t *testing.T) {
	ctx := context.Background()
	_, conn, cancel := createPair(t, grpc.WithIdleTimeout(time.Second))
	defer cancel()
	cli := NewGreeterClient(conn)
	_, err := cli.SayHello(ctx, &HelloRequest{})
	require.NoError(t, err)
	time.Sleep(time.Hour)
}

func Test_1(t *testing.T) {
	ctx := context.Background()
	_, conn, cancel := createPair(t)
	defer cancel()

	cli := NewGreeterClient(conn)
	_, err := cli.SayHello(ctx, &HelloRequest{})
	require.NoError(t, err)
}

func Test_MetaData(t *testing.T) {
	ctx := context.Background()
	_, conn, cancel := createPair(t)
	defer cancel()

	cli := NewGreeterClient(conn)
	var header metadata.MD
	var trailerHeader metadata.MD
	// pass meta data to server
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "bearer token123")
	_, err := cli.SayHello(ctx, &HelloRequest{}, grpc.Trailer(&trailerHeader), grpc.Header(&header))
	require.NoError(t, err)
}

func Test_ReturnTwoNil(t *testing.T) {
	ctx := context.Background()
	_, conn, cancel := createPair(t)
	defer cancel()

	cli := NewGreeterClient(conn)
	resp, err := cli.SayHelloReturnNil(ctx, &HelloRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_RingHashBalancer(t *testing.T) {
	ctx := context.Background()
	// must set environment GRPC_EXPERIMENTAL_RING_HASH_SET_REQUEST_HASH_KEY
	serviceConfig := `{
		"loadBalancingConfig": [ 
			{ 
				"ring_hash_experimental": {
					"minRingSize": 1024,
					"maxRingSize": 4096
				}
			}
		]
	}`
	_, conn, cancel := createPair(t, grpc.WithDefaultServiceConfig(serviceConfig))
	defer cancel()

	cli := NewGreeterClient(conn)
	resp, err := cli.SayHello(ctx, &HelloRequest{})
	require.NoError(t, err)
	_ = resp
}
