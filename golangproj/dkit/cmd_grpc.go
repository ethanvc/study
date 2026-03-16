package dkit

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ethanvc/study/golangproj/xobs"
	"github.com/spf13/cobra"
	"sort"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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
	case "show-method":
		return queryShowMethod(req)
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
	GetFileDescriptorsBySymbol(ctx context.Context, symbol string) ([]*descriptorpb.FileDescriptorProto, error)
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

func (c *reflectionClientV1) GetFileDescriptorsBySymbol(ctx context.Context, symbol string) ([]*descriptorpb.FileDescriptorProto, error) {
	stream, err := reflectionv1.NewServerReflectionClient(c.cc).ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(&reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: symbol,
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
	return parseFileDescriptors(fdResp.GetFileDescriptorProto())
}

func (c *reflectionClientV1) ListMethods(ctx context.Context, service string) ([]string, error) {
	fds, err := c.GetFileDescriptorsBySymbol(ctx, service)
	if err != nil {
		return nil, err
	}
	return extractMethods(fds, service)
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

func (c *reflectionClientV1Alpha) GetFileDescriptorsBySymbol(ctx context.Context, symbol string) ([]*descriptorpb.FileDescriptorProto, error) {
	stream, err := reflectionv1alpha.NewServerReflectionClient(c.cc).ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}
	if err := stream.Send(&reflectionv1alpha.ServerReflectionRequest{
		MessageRequest: &reflectionv1alpha.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: symbol,
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
	return parseFileDescriptors(fdResp.GetFileDescriptorProto())
}

func (c *reflectionClientV1Alpha) ListMethods(ctx context.Context, service string) ([]string, error) {
	fds, err := c.GetFileDescriptorsBySymbol(ctx, service)
	if err != nil {
		return nil, err
	}
	return extractMethods(fds, service)
}

func parseFileDescriptors(rawDescs [][]byte) ([]*descriptorpb.FileDescriptorProto, error) {
	var fds []*descriptorpb.FileDescriptorProto
	for _, raw := range rawDescs {
		fd := &descriptorpb.FileDescriptorProto{}
		if err := proto.Unmarshal(raw, fd); err != nil {
			return nil, fmt.Errorf("unmarshal file descriptor: %w", err)
		}
		fds = append(fds, fd)
	}
	return fds, nil
}

func extractMethods(fds []*descriptorpb.FileDescriptorProto, service string) ([]string, error) {
	for _, fd := range fds {
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

func queryShowMethod(req *GrpcMainReq) error {
	ctx := context.Background()
	rc, err := NewReflectionClient(ctx, &GrpcClientConfig{Host: req.Host})
	if err != nil {
		return err
	}
	defer rc.Close()

	svcName, methodName, err := parseMethodPath(req.Method)
	if err != nil {
		return err
	}

	fds, err := rc.GetFileDescriptorsBySymbol(ctx, svcName)
	if err != nil {
		return err
	}

	md, err := findMethodDescriptor(fds, svcName, methodName)
	if err != nil {
		return err
	}

	inputFQ := md.GetInputType()
	outputFQ := md.GetOutputType()
	fmt.Printf("rpc %s(%s%s) returns (%s%s)\n\n",
		md.GetName(),
		streamPrefix(md.GetClientStreaming()),
		shortTypeName(inputFQ),
		streamPrefix(md.GetServerStreaming()),
		shortTypeName(outputFQ),
	)

	resolver := &messageResolver{
		rc:      rc,
		ctx:     ctx,
		fdMap:   make(map[string]*descriptorpb.FileDescriptorProto),
		printed: make(map[string]bool),
	}
	for _, fd := range fds {
		resolver.fdMap[fd.GetName()] = fd
	}
	resolver.resolveAndPrint(inputFQ)
	resolver.resolveAndPrint(outputFQ)
	return nil
}

type messageResolver struct {
	rc      ReflectionClient
	ctx     context.Context
	fdMap   map[string]*descriptorpb.FileDescriptorProto
	printed map[string]bool
}

func (r *messageResolver) resolveAndPrint(typeFQ string) {
	normalized := strings.TrimPrefix(typeFQ, ".")
	if r.printed[normalized] {
		return
	}
	r.printed[normalized] = true

	msg := findMessageInFdMap(r.fdMap, typeFQ)
	if msg == nil {
		moreFds, err := r.rc.GetFileDescriptorsBySymbol(r.ctx, normalized)
		if err == nil {
			for _, fd := range moreFds {
				r.fdMap[fd.GetName()] = fd
			}
			msg = findMessageInFdMap(r.fdMap, typeFQ)
		}
	}
	if msg == nil {
		return
	}

	printMessageDescriptor(msg, typeFQ)

	for _, f := range msg.GetField() {
		if f.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
			r.resolveAndPrint(f.GetTypeName())
		}
	}
}

func parseMethodPath(method string) (service, methodName string, err error) {
	method = strings.TrimPrefix(method, "/")
	idx := strings.LastIndex(method, "/")
	if idx < 0 {
		return "", "", fmt.Errorf("invalid method path %q, expected format: package.Service/Method", method)
	}
	return method[:idx], method[idx+1:], nil
}

func findMethodDescriptor(fds []*descriptorpb.FileDescriptorProto, service, method string) (*descriptorpb.MethodDescriptorProto, error) {
	for _, fd := range fds {
		for _, svc := range fd.GetService() {
			fqn := fd.GetPackage() + "." + svc.GetName()
			if fqn != service {
				continue
			}
			for _, m := range svc.GetMethod() {
				if m.GetName() == method {
					return m, nil
				}
			}
			return nil, fmt.Errorf("method %q not found in service %q", method, service)
		}
	}
	return nil, fmt.Errorf("service %q not found in file descriptors", service)
}

func findMessageInFdMap(fdMap map[string]*descriptorpb.FileDescriptorProto, fqn string) *descriptorpb.DescriptorProto {
	fqn = strings.TrimPrefix(fqn, ".")
	for _, fd := range fdMap {
		pkg := fd.GetPackage()
		for _, msg := range fd.GetMessageType() {
			if found := findNestedMessage(msg, pkg+"."+msg.GetName(), fqn); found != nil {
				return found
			}
		}
	}
	return nil
}

func findNestedMessage(msg *descriptorpb.DescriptorProto, prefix, target string) *descriptorpb.DescriptorProto {
	if prefix == target {
		return msg
	}
	for _, nested := range msg.GetNestedType() {
		if found := findNestedMessage(nested, prefix+"."+nested.GetName(), target); found != nil {
			return found
		}
	}
	return nil
}

func printMessageDescriptor(msg *descriptorpb.DescriptorProto, fqn string) {
	fmt.Printf("message %s {\n", shortTypeName(fqn))
	for _, f := range msg.GetField() {
		label := ""
		if f.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
			label = "repeated "
		}
		fmt.Printf("  %s%s %s = %d;\n", label, protoFieldTypeName(f), f.GetName(), f.GetNumber())
	}
	fmt.Println("}")
	fmt.Println()
}

func shortTypeName(fqn string) string {
	fqn = strings.TrimPrefix(fqn, ".")
	if idx := strings.LastIndex(fqn, "."); idx >= 0 {
		return fqn[idx+1:]
	}
	return fqn
}

func streamPrefix(streaming bool) string {
	if streaming {
		return "stream "
	}
	return ""
}

func protoFieldTypeName(f *descriptorpb.FieldDescriptorProto) string {
	switch f.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE,
		descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return shortTypeName(f.GetTypeName())
	default:
		return strings.TrimPrefix(strings.ToLower(f.GetType().String()), "type_")
	}
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
	var header metadata.MD
	err = cc.Invoke(
		context.Background(), req.Method, body, &resp,
		grpc.ForceCodec(NewRawCodec(req.SubType)),
		grpc.Header(&header),
	)
	if err != nil {
		return err
	}

	printMetadata(header)

	if len(resp) == 0 {
		fmt.Fprintln(os.Stderr, "(empty response)")
		return nil
	}
	_, err = os.Stdout.Write(resp)
	return err
}

func printMetadata(md metadata.MD) {
	if len(md) == 0 {
		return
	}
	keys := make([]string, 0, len(md))
	for k := range md {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range md[k] {
			fmt.Fprintf(os.Stderr, "%s: %s\n", k, v)
		}
	}
	fmt.Fprintln(os.Stderr)
}

type GrpcClientConfig struct {
	Host string
}

func NewGrpcClient(conf *GrpcClientConfig) (*grpc.ClientConn, error) {
	cc, err := grpc.NewClient(conf.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return cc, err
}
