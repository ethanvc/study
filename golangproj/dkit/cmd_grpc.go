package dkit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	reflectionv1 "google.golang.org/grpc/reflection/grpc_reflection_v1"
	reflectionv1alpha "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

// RawCodec is a gRPC codec that passes bytes through without marshaling.
type RawCodec struct {
	name string
}

func NewRawCodec(name string) *RawCodec {
	return &RawCodec{name: name}
}

func (c *RawCodec) Marshal(v any) ([]byte, error) {
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	default:
		return nil, fmt.Errorf("raw codec: unsupported marshal type %T", v)
	}
}

func (c *RawCodec) Unmarshal(data []byte, v any) error {
	switch ptr := v.(type) {
	case *[]byte:
		*ptr = data
		return nil
	default:
		return fmt.Errorf("raw codec: expect *[]byte, received %T", v)
	}
}

func (c *RawCodec) Name() string {
	return c.name
}

func AddGrpcCmd(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use: "grpc",
	}
	host := cmd.Flags().String("host", "127.0.0.1:8888", "server instance address")
	method := cmd.Flags().String("method", "/helloworld.Greeter/SayHello", "method name")
	body := cmd.Flags().String("body", "", "request content")
	subType := cmd.Flags().String("sub-type", "", "content sub-type (e.g. proto, json)")
	query := cmd.Flags().String("query", "", "query type")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return GrpcMain(&GrpcMainReq{
			Host:    *host,
			Body:    *body,
			Method:  *method,
			SubType: *subType,
			Query:   *query,
		})
	}
	rootCmd.AddCommand(cmd)
}

type GrpcMainReq struct {
	Host    string
	Body    string
	Method  string
	SubType string
	Query   string
}

func resolveBody(body string) ([]byte, error) {
	if strings.HasPrefix(body, "@") {
		return os.ReadFile(body[1:])
	}
	return []byte(body), nil
}

func GrpcMain(req *GrpcMainReq) error {
	if req.Query != "" {
		return queryByReflect(req)
	}
	return sendRequest(req)
}

func queryByReflect(req *GrpcMainReq) error {
	switch req.Query {
	case "list-svr":
		return querySvrList(req)
	}
	return errors.New("invalid query value")
}

func querySvrList(req *GrpcMainReq) error {
	cc, err := NewGrpcClient(&GrpcClientConfig{
		Host: req.Host,
	})
	if err != nil {
		return fmt.Errorf("dial server: %w", err)
	}
	defer cc.Close()

	names, err := listServicesV1(cc)
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.Unimplemented {
			names, err = listServicesV1Alpha(cc)
		}
	}
	if err != nil {
		return err
	}
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

func listServicesV1(cc *grpc.ClientConn) ([]string, error) {
	stream, err := reflectionv1.NewServerReflectionClient(cc).ServerReflectionInfo(context.Background())
	if err != nil {
		return nil, err
	}
	if err := stream.Send(&reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_ListServices{},
	}); err != nil {
		return nil, err
	}
	stream.CloseSend()
	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	list := resp.GetListServicesResponse()
	if list == nil {
		return nil, fmt.Errorf("unexpected response: %v", resp.GetMessageResponse())
	}
	var names []string
	for _, svc := range list.GetService() {
		names = append(names, svc.GetName())
	}
	return names, nil
}

func listServicesV1Alpha(cc *grpc.ClientConn) ([]string, error) {
	stream, err := reflectionv1alpha.NewServerReflectionClient(cc).ServerReflectionInfo(context.Background())
	if err != nil {
		return nil, err
	}
	if err := stream.Send(&reflectionv1alpha.ServerReflectionRequest{
		MessageRequest: &reflectionv1alpha.ServerReflectionRequest_ListServices{},
	}); err != nil {
		return nil, err
	}
	stream.CloseSend()
	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	list := resp.GetListServicesResponse()
	if list == nil {
		return nil, fmt.Errorf("unexpected response: %v", resp.GetMessageResponse())
	}
	var names []string
	for _, svc := range list.GetService() {
		names = append(names, svc.GetName())
	}
	return names, nil
}

func sendRequest(req *GrpcMainReq) error {
	body, err := resolveBody(req.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	cc, err := NewGrpcClient(&GrpcClientConfig{
		Host: req.Host,
	})
	if err != nil {
		return fmt.Errorf("dial server: %w", err)
	}
	defer cc.Close()

	var resp []byte
	err = cc.Invoke(
		context.Background(), req.Method, body, &resp,
		grpc.ForceCodec(NewRawCodec(req.SubType)),
	)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(resp)
	return err
}

type GrpcClientConfig struct {
	Host string
}

func NewGrpcClient(conf *GrpcClientConfig) (*grpc.ClientConn, error) {
	cc, err := grpc.NewClient(conf.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return cc, err
}
