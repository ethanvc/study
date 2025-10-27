package golangapirouter

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func NewExecutionStateFromHttpRequest(rule *Rule, req *http.Request) *ExecutionState {
	state := &ExecutionState{
		protoSpec: NewHttpProtocolSpec(req),
		header:    req.Header,
		body: &HttpBody{
			req: req,
		},
		rule: rule,
	}
	return state
}

type HttpBody struct {
	req *http.Request
}

func (body *HttpBody) Get(key string) string {
	return ""
}

type HttpProtocolSpec struct {
	Method     string
	Url        *url.URL
	Host       string
	Proto      string
	ProtoMajor int
	ProtoMinor int
	queries    url.Values
}

func NewHttpProtocolSpec(r *http.Request) *HttpProtocolSpec {
	return &HttpProtocolSpec{
		Method:     r.Method,
		Url:        r.URL,
		Host:       r.Host,
		Proto:      r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
	}
}

func (spec *HttpProtocolSpec) Get(key string) string {
	switch key {
	case "method":
		return spec.Method
	case "url":
		return spec.Url.String()
	case "path":
		return spec.Url.Path
	case "host":
		return spec.Host
	case "proto":
		return spec.Proto
	case "proto_major":
		return strconv.Itoa(spec.ProtoMajor)
	case "proto_minor":
		return strconv.Itoa(spec.ProtoMinor)
	}
	const prefixQuery = "query."
	if strings.HasPrefix(key, prefixQuery) {
		key = key[len(prefixQuery):]
		if spec.queries == nil {
			spec.queries = spec.Url.Query()
		}
		return spec.queries.Get(key)
	}
	return ""
}
