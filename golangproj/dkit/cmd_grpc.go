package dkit

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RawCodec is a gRPC codec that passes bytes through without marshaling.
// Name() returns "proto" so the server treats it as a normal protobuf request.
type RawCodec struct{}

func (RawCodec) Marshal(v any) ([]byte, error) {
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	default:
		return nil, fmt.Errorf("raw codec: unsupported marshal type %T", v)
	}
}

func (RawCodec) Unmarshal(data []byte, v any) error {
	switch ptr := v.(type) {
	case *[]byte:
		*ptr = data
		return nil
	default:
		return fmt.Errorf("raw codec: unsupported unmarshal type %T", v)
	}
}

func (RawCodec) Name() string {
	return "proto"
}

func AddGrpcCmd(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:  "grpc [method]",
		Args: cobra.ExactArgs(1),
	}
	host := cmd.Flags().String("host", "127.0.0.1:8888", "server instance address")
	body := cmd.Flags().String("body", "", "request content")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return SendGrpcRequest(&SendGrpcRequestReq{
			Host:   *host,
			Body:   *body,
			Method: args[0],
		})
	}
	rootCmd.AddCommand(cmd)
}

type SendGrpcRequestReq struct {
	Host   string
	Body   string
	Method string
}

func SendGrpcRequest(req *SendGrpcRequestReq) error {
	cc, err := grpc.NewClient(req.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("dial server: %w", err)
	}
	defer cc.Close()

	var resp []byte
	method := "/" + req.Method
	err = cc.Invoke(context.Background(), method, []byte(req.Body), &resp, grpc.ForceCodec(RawCodec{}))
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(resp)
	return err
}
