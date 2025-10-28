package golangapirouter

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const ExprValueTrue = "true"
const ExprValueFalse = "false"

type builtinIn struct {
}

func (builtinIn) Name() string {
	return "_in"
}

func (builtinIn) ValidateParamRef(paramRef []ParamRef) error {
	if len(paramRef) < 2 {
		return fmt.Errorf("_in: expecting at least 2 params")
	}
	return nil
}

func (builtinIn) Call(ctx context.Context, execCtx *ExecutionState, args []string) (string, error) {
	if slices.Contains(args[1:], args[0]) {
		return ExprValueTrue, nil
	}
	return ExprValueFalse, nil
}

type builtinHasPrefix struct {
}

func (builtinHasPrefix) Name() string {
	return "_has_prefix"
}

func (builtinHasPrefix) ValidateParamRef(paramRef []ParamRef) error {
	if len(paramRef) < 2 {
		return fmt.Errorf("_has_prefix: expecting at least 2 params")
	}
	return nil
}

func (builtinHasPrefix) Call(ctx context.Context, execCtx *ExecutionState, args []string) (string, error) {
	for _, arg := range args[1:] {
		if strings.HasPrefix(args[0], arg) {
			return ExprValueTrue, nil
		}
	}
	return ExprValueFalse, nil
}

type WriteLogReq struct {
	UnitId     int      `json:"unit_id"`
	BusinessId int      `json:"business_id"`
	LogIds     []string `json:"log_ids"`
}

type ReadLogReq struct {
	BusinessId int    `json:"business_id"`
	LogId      string `json:"log_id"`
}

var ErrRoutingLogNotExist = errors.New("routing log does not exist")

type IRoutingLog interface {
	WriteLog(ctx context.Context, req *WriteLogReq) (int, error)
	ReadLog(ctx context.Context, req *ReadLogReq) (int, error)
}

type BuiltinRouteByRoutingLog struct {
	log IRoutingLog
}

func NewRouteByRoutingLog(log IRoutingLog) *BuiltinRouteByRoutingLog {
	return &BuiltinRouteByRoutingLog{
		log: log,
	}
}

func (f *BuiltinRouteByRoutingLog) Name() string {
	return "_route_by_routing_log"
}

func (f *BuiltinRouteByRoutingLog) VerifyParamRefs(ctx context.Context, paramRef []ParamRef) error {
	return nil
}

func (f *BuiltinRouteByRoutingLog) Call(ctx context.Context, execCtx *ExecutionState, args []string) (string, error) {
	return "", nil
}

type builtinConsult struct{}

func (builtinConsult) Name() string {
	return "_consult"
}

func (builtinConsult) ValidateParamRef(paramRefs []ParamRef) error {
	return nil
}

func (builtinConsult) Call(ctx context.Context, cxCtx *ExecutionState, args []string) (string, error) {
	return "", nil
}

type builtinRouteConsultAndRetry struct {
}

func (builtinRouteConsultAndRetry) Name() string {
	return "_route_consult_and_retry"
}
func (builtinRouteConsultAndRetry) ValidateParamRef(paramRefs []ParamRef) error {
	return nil
}
func (builtinRouteConsultAndRetry) Call(ctx context.Context, cxCtx *ExecutionState, args []string) (string, error) {
	return "", nil
}

type builtinRouteAny struct {
}

func (builtinRouteAny) Name() string {
	return "_route_any"
}
func (builtinRouteAny) ValidateParamRef(paramRef []ParamRef) error {
	return nil
}
func (builtinRouteAny) Call(ctx context.Context, cxCtx *ExecutionState, args []string) (string, error) {
	return "", nil
}

type builtinRouteRoutingLog struct {
}

func (builtinRouteRoutingLog) Name() string {
	return "_route_routing_log"
}
func (builtinRouteRoutingLog) ValidateParamRef(paramRef []ParamRef) error {
	return nil
}
func (builtinRouteRoutingLog) Call(ctx context.Context, cxCtx *ExecutionState, args []string) (string, error) {
	return "", nil
}
