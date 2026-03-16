package dkit

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ethanvc/study/golangproj/xobs"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	reflectionv1 "google.golang.org/grpc/reflection/grpc_reflection_v1"
	reflectionv1alpha "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
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
	return xobs.New(codes.InvalidArgument, "InvalidQueryValue").SetMsg("invalid query value")
}

func queryMethodList(req *GrpcMainReq) error {
	ctx := context.Background()
	rc, err := NewReflectionClient(ctx, &GrpcClientConfig{Host: req.Host})
	if err != nil {
		return err
	}
	defer rc.Close()
	methods, err := rc.ListMethods(ctx, req.Svr)
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
	Close() error
}

// NewReflectionClient tries v1 and v1alpha with independent connections,
// returns the first that succeeds, or both errors if both fail.
func NewReflectionClient(ctx context.Context, conf *GrpcClientConfig) (ReflectionClient, error) {
	v1, err1 := newReflectionClientV1(conf)
	if err1 == nil {
		_, err1 = v1.ListServices(ctx)
		if err1 == nil {
			return v1, nil
		}
		v1.Close()
	}

	v1alpha, err2 := newReflectionClientV1Alpha(conf)
	if err2 == nil {
		_, err2 = v1alpha.ListServices(ctx)
		if err2 == nil {
			return v1alpha, nil
		}
		v1alpha.Close()
	}

	return nil, fmt.Errorf("v1: %w; v1alpha: %w", err1, err2)
}

type reflectionClientV1 struct {
	cc *grpc.ClientConn
}

func newReflectionClientV1(conf *GrpcClientConfig) (*reflectionClientV1, error) {
	cc, err := NewGrpcClient(conf)
	if err != nil {
		return nil, err
	}
	return &reflectionClientV1{cc: cc}, nil
}

func (c *reflectionClientV1) Close() error { return c.cc.Close() }

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

func newReflectionClientV1Alpha(conf *GrpcClientConfig) (*reflectionClientV1Alpha, error) {
	cc, err := NewGrpcClient(conf)
	if err != nil {
		return nil, err
	}
	return &reflectionClientV1Alpha{cc: cc}, nil
}

func (c *reflectionClientV1Alpha) Close() error { return c.cc.Close() }

func (c *reflectionClientV1Alpha) ListServices(ctx context.Context) ([]string, error) {
	stream, err := reflectionv1alpha.NewServerReflectionClient(c.cc).ServerReflectionInfo(ctx)
	if err != nil {
		return nil, xobs.New(codes.Internal, "CallServerReflectionInfoErr").SetMsg(err.Error())
	}
	if err = stream.Send(&reflectionv1alpha.ServerReflectionRequest{
		MessageRequest: &reflectionv1alpha.ServerReflectionRequest_ListServices{},
	}); err != nil {
		return nil, xobs.New(codes.Internal, "CallStreamSendErr").SetMsg(err.Error())
	}
	stream.CloseSend()
	resp, err := stream.Recv()
	if err != nil {
		return nil, xobs.New(codes.Internal, "CallStreamRecvErr").SetMsg(err.Error())
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
	ctx := context.Background()
	rc, err := NewReflectionClient(ctx, &GrpcClientConfig{Host: req.Host})
	if err != nil {
		return err
	}
	defer rc.Close()
	names, err := rc.ListServices(ctx)
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
