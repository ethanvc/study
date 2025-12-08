package dkit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func AddGrpcCmd(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use: "grpc",
	}
	target := cmd.Flags().String("target", "", "target name")
	method := cmd.Flags().String("method", "", "grpc full method")
	reqBody := cmd.Flags().String("body", "", "body content")
	contentSubtype := cmd.Flags().String("content-subtype", "", "")
	insecureFlag := cmd.Flags().Bool("insecure", true, "grpc.WithTransportCredentials(insecure.NewCredentials())")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return SendGrpcRequest(context.Background(), &SendGrpcRequestReq{
			Target:         *target,
			Method:         *method,
			ContentSubtype: *contentSubtype,
			Insecure:       *insecureFlag,
			ReqBody:        *reqBody,
		})
	}
	rootCmd.AddCommand(cmd)
}

type SendGrpcRequestReq struct {
	Target         string
	Method         string
	ContentSubtype string
	Insecure       bool
	ReqBody        string
}

func SendGrpcRequest(ctx context.Context, req *SendGrpcRequestReq) error {
	var dialOpts []grpc.DialOption
	if req.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	dialOpts = append(dialOpts, grpc.WithBlock())
	conn, err := grpc.NewClient(req.Target, dialOpts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 准备调用上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	var resp string

	err = conn.Invoke(ctx, req.Method, req.ReqBody, &resp, grpc.CallContentSubtype(req.ContentSubtype))
	if err != nil {
		return err
	}
	fmt.Printf("resp: %s\n", resp)
	return nil
}

type JsonCodec struct {
}

func (_ JsonCodec) Name() string {
	return "fulljson"
}

func (_ JsonCodec) Marshal(v interface{}) ([]byte, error) {
	switch realV := v.(type) {
	case []byte:
		return realV, nil
	case string:
		return []byte(realV), nil
	default:
		return nil, errors.New("UnableToMarshalGrpcPackage")
	}
}

func (_ JsonCodec) Unmarshal(data []byte, v interface{}) error {
	switch realV := v.(type) {
	case *[]byte:
		*realV = data
		return nil
	case *string:
		*realV = string(data)
	default:
		return errors.New("UnableToUnmarshalGrpcPackage")
	}
	return nil
}
