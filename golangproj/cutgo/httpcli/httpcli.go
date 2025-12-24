package httpcli

import (
	"net/http"

	"github.com/ethanvc/study/golangproj/cutgo"
)
import "context"

var defaultCli = NewHttpClient()

func Call(ctx context.Context, method string, req any, resp any, info *CallInfo) error {
	return defaultCli.Call(ctx, method, req, resp, info)
}

type HttpClient struct {
	handler *cutgo.Handler[CallInfo]
}

func NewHttpClient() *HttpClient {
	cli := &HttpClient{}
	cli.init()
	return cli
}

func (cli *HttpClient) init() {
	cli.handler = cutgo.NewHandler(cli.do)
}

func (cli *HttpClient) do(ctx context.Context, req, resp any, info *CallInfo) error {
	return nil
}

func (cli *HttpClient) Call(ctx context.Context, url string, req, resp any, info *CallInfo) error {
	return cli.handler.Call(ctx, url, req, resp, info)
}

type CallInfo struct {
	Method string
	Header http.Header
}
