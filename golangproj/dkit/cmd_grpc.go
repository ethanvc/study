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
	"google.golang.org/protobuf/proto"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

func AddGrpcCmd(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use: "grpc",
	}
	host := cmd.Flags().String("host", "127.0.0.1:8888", "server instance address")
	method := cmd.Flags().String("method", "/helloworld.Greeter/SayHello", "method name")
	body := cmd.Flags().String("body", "", "request content")
	subType := cmd.Flags().String("sub-type", "", "content sub-type (e.g. proto, json)")
	query := cmd.Flags().String("query", "", "query type")
	svr := cmd.Flags().String("svr", "", "server name")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return GrpcMain(&GrpcMainReq{
			Host:    *host,
			Body:    *body,
			Method:  *method,
			SubType: *subType,
			Query:   *query,
			Svr:     *svr,
		})
	}
	rootCmd.AddCommand(cmd)
}

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

type GrpcMainReq struct {
	Host    string
	Body    string
	Method  string
	SubType string
	Query   string
	Svr     string
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
	case "list-method":
		return queryMethodList(req)
	}
	return errors.New("invalid query value")
}

func queryMethodList(req *GrpcMainReq) error {
	cc, err := NewGrpcClient(&GrpcClientConfig{Host: req.Host})
	if err != nil {
		return fmt.Errorf("dial server: %w", err)
	}
	defer cc.Close()

	rc, err := NewReflectionClient(cc)
	if err != nil {
		return err
	}
	methods, err := rc.ListMethods(context.Background(), req.Svr)
	if err != nil {
		return err
	}
	for _, m := range methods {
		fmt.Println(m)
	}
	return nil
}

type ReflectionClient interface {
	ListServices(ctx context.Context) ([]string, error)
	ListMethods(ctx context.Context, service string) ([]string, error)
}

// NewReflectionClient tries v1 first; on Unimplemented falls back to v1alpha.
func NewReflectionClient(cc *grpc.ClientConn) (ReflectionClient, error) {
	v1 := &reflectionClientV1{cc: cc}
	_, err := v1.ListServices(context.Background())
	if err == nil {
		return v1, nil
	}
	if s, ok := status.FromError(err); ok && s.Code() == codes.Unimplemented {
		return &reflectionClientV1Alpha{cc: cc}, nil
	}
	return nil, err
}

type reflectionClientV1 struct {
	cc *grpc.ClientConn
}

func (c *reflectionClientV1) ListServices(ctx context.Context) ([]string, error) {
	stream, err := reflectionv1.NewServerReflectionClient(c.cc).ServerReflectionInfo(ctx)
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

func (c *reflectionClientV1) ListMethods(ctx context.Context, service string) ([]string, error) {
	stream, err := reflectionv1.NewServerReflectionClient(c.cc).ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(&reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: service,
		},
	}); err != nil {
		return nil, err
	}
	stream.CloseSend()
	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	fdResp := resp.GetFileDescriptorResponse()
	if fdResp == nil {
		return nil, fmt.Errorf("unexpected response: %v", resp.GetMessageResponse())
	}
	return extractMethods(fdResp.GetFileDescriptorProto(), service)
}

type reflectionClientV1Alpha struct {
	cc *grpc.ClientConn
}

func (c *reflectionClientV1Alpha) ListServices(ctx context.Context) ([]string, error) {
	stream, err := reflectionv1alpha.NewServerReflectionClient(c.cc).ServerReflectionInfo(ctx)
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

func (c *reflectionClientV1Alpha) ListMethods(ctx context.Context, service string) ([]string, error) {
	stream, err := reflectionv1alpha.NewServerReflectionClient(c.cc).ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(&reflectionv1alpha.ServerReflectionRequest{
		MessageRequest: &reflectionv1alpha.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: service,
		},
	}); err != nil {
		return nil, err
	}
	stream.CloseSend()
	resp, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	fdResp := resp.GetFileDescriptorResponse()
	if fdResp == nil {
		return nil, fmt.Errorf("unexpected response: %v", resp.GetMessageResponse())
	}
	return extractMethods(fdResp.GetFileDescriptorProto(), service)
}

func extractMethods(rawDescs [][]byte, service string) ([]string, error) {
	for _, raw := range rawDescs {
		fd := &descriptorpb.FileDescriptorProto{}
		if err := proto.Unmarshal(raw, fd); err != nil {
			return nil, fmt.Errorf("unmarshal file descriptor: %w", err)
		}
		for _, svc := range fd.GetService() {
			fqn := fd.GetPackage() + "." + svc.GetName()
			if fqn != service {
				continue
			}
			var methods []string
			for _, m := range svc.GetMethod() {
				methods = append(methods, fqn+"/"+m.GetName())
			}
			return methods, nil
		}
	}
	return nil, fmt.Errorf("service %q not found in file descriptors", service)
}

func querySvrList(req *GrpcMainReq) error {
	cc, err := NewGrpcClient(&GrpcClientConfig{Host: req.Host})
	if err != nil {
		return fmt.Errorf("dial server: %w", err)
	}
	defer cc.Close()

	rc, err := NewReflectionClient(cc)
	if err != nil {
		return err
	}
	names, err := rc.ListServices(context.Background())
	if err != nil {
		return err
	}
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
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
