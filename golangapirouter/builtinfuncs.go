package golangapirouter

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

const ExprValueTrue = "true"
const ExprValueFalse = "false"

type builtinIn struct {
}

func (f *builtinIn) Name() string {
	return "_in"
}

func (f *builtinIn) VerifyParamRefs(ctx context.Context, paramRef []ParamRef) error {
	if len(paramRef) < 2 {
		return fmt.Errorf("_in: expecting at least 2 params")
	}
	return nil
}

func (f *builtinIn) Call(ctx context.Context, execCtx *RuleExecutionState, args []string) ([]string, error) {
	if slices.Contains(args[1:], args[0]) {
		return []string{ExprValueTrue}, nil
	}
	return []string{ExprValueFalse}, nil
}

type builtinHasPrefix struct {
}

func (f *builtinHasPrefix) Name() string {
	return "_has_prefix"
}

func (f *builtinHasPrefix) VerifyParamRefs(ctx context.Context, paramRef []ParamRef) error {
	if len(paramRef) < 2 {
		return fmt.Errorf("_has_prefix: expecting at least 2 params")
	}
	return nil
}

func (f *builtinHasPrefix) Call(ctx context.Context, execCtx *RuleExecutionState, args []string) ([]string, error) {
	for _, arg := range args[1:] {
		if strings.HasPrefix(args[0], arg) {
			return []string{ExprValueTrue}, nil
		}
	}
	return []string{ExprValueFalse}, nil
}

type builtinRouteByRoutingLog struct {
}

func (f *builtinRouteByRoutingLog) Name() string {
	return "_route_by_routing_log"
}

func (f *builtinRouteByRoutingLog) VerifyParamRefs(ctx context.Context, paramRef []ParamRef) error {
	return nil
}

func (f *builtinRouteByRoutingLog) Call(ctx context.Context, execCtx *RuleExecutionState, args []string) (string, error) {
	return "", nil
}
