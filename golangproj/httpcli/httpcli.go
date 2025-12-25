package httpcli

import (
	"context"
	"net/http"
	"time"
)

var sDefaultClient = &Client{}

func GetDefault() *Client {
	return sDefaultClient
}

func Do(ctx context.Context, url string, req, resp any, opts *Options) error {
	return GetDefault().Do(ctx, url, req, resp, opts)
}

type Client struct {
	Serializer   Serializer
	Timeout      time.Duration
	Interceptors []Interceptor
}

func (cli *Client) Do(ctx context.Context, url string, req, resp any, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	next := Next{
		interceptors: cli.Interceptors,
		handler:      cli.handle,
	}
	return next.Next(ctx, url, req, resp, opts)
}

func (cli *Client) handle(ctx context.Context, url string, req, resp any, opts *Options) error {
	return nil
}

type Interceptor func(ctx context.Context, url string, req, resp any, opts *Options, next Next) error

type Next struct {
	i            int
	interceptors []Interceptor
	handler      func(ctx context.Context, url string, req, resp any, opts *Options) error
}

func (n Next) Next(ctx context.Context, url string, req, resp any, opts *Options) error {
	if n.i >= len(n.interceptors) {
		return n.handler(ctx, url, req, resp, opts)
	}
	newN := n
	newN.i++
	return n.interceptors[n.i](ctx, url, req, resp, opts, newN)
}

type Options struct {
	Method     string
	Timeout    time.Duration
	Serializer Serializer
}

type Serializer interface {
	Marshal(ctx context.Context, v any, opts *Options) (string, []byte, error)
	Unmarshal(ctx context.Context, httpResp *http.Response, resp any, opts *Options) error
}
