package httpcli

import (
	"context"
	"net/http"
	"time"
)

var DefaultSerializer = &AutoSerializer{}

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
	ctx, cancel := cli.handleTimeout(ctx, opts.Timeout)
	if cancel != nil {
		defer cancel()
	}
	next := Next{
		interceptors: cli.Interceptors,
		handler:      cli.handle,
	}
	return next.Next(ctx, url, req, resp, opts)
}

func (cli *Client) handle(ctx context.Context, url string, req, resp any, opts *Options) error {
	serializer := cli.getSerializer(opts)
	contentType, reqBody, err := serializer.Marshal(ctx, req, opts)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, opts.GetMethod(), url, reqBody)
	if err != nil {
		return err
	}
	if len(opts.Header) > 0 {
		httpReq.Header = opts.Header
	}
	if contentType != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	httpResp, err := cli.getHttpClient().Do(httpReq)
	if err != nil {
		return err
	}
	opts.StatusCode = httpResp.StatusCode
	opts.RespHeader = httpResp.Header
	err = serializer.Unmarshal(ctx, httpResp, resp, opts)
	if err != nil {
		return err
	}
	return nil
}

func (cli *Client) handleTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	realTimeout := timeout
	if realTimeout == 0 {
		realTimeout = cli.Timeout
	}
	if realTimeout == 0 {
		return ctx, nil
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return context.WithTimeout(ctx, realTimeout)
	}
	existingTimeout := deadline.Sub(time.Now())
	if existingTimeout < realTimeout {
		return ctx, nil
	}
	return context.WithTimeout(ctx, realTimeout)
}

func (cli *Client) getHttpClient() *http.Client {
	return http.DefaultClient
}

func (cli *Client) getSerializer(opts *Options) Serializer {
	if opts.Serializer != nil {
		return opts.Serializer
	}
	if cli.Serializer != nil {
		return cli.Serializer
	}
	return DefaultSerializer
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
	Header     http.Header
	Timeout    time.Duration
	Serializer Serializer

	// output fields
	StatusCode int
	RespBody   []byte
	RespHeader http.Header
}

func (opts *Options) GetMethod() string {
	if opts.Method != "" {
		return opts.Method
	}
	return http.MethodPost
}
