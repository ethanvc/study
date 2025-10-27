package golangapirouter

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type RuleConfig struct {
	Api     string          `json:"api"`
	Actions []*ActionConfig `json:"actions"`
}

func (ruleConf *RuleConfig) Validate() error {
	if ruleConf.Api == "" {
		return errors.New("api is required")
	}
	for i, action := range ruleConf.Actions {
		err := action.Validate()
		if err != nil {
			return fmt.Errorf("actions[%d] invalid: %w", i, err)
		}
	}
	return nil
}

type ActionConfig struct {
	Name          string   `json:"name"`
	Args          []string `json:"args,omitempty"`
	ComputeKey    string   `json:"compute_key,omitempty"`
	Terminal      *bool    `json:"terminal,omitempty"`
	Condition     string   `json:"condition,omitempty"`
	ConditionArgs []string `json:"condition_args,omitempty"`
	ConditionAnd  []Expr   `json:"condition_and,omitempty"`
	ConditionOr   []Expr   `json:"condition_or,omitempty"`
}

func (action *ActionConfig) Validate() error {
	return nil
}

type Expr struct {
	Name string   `json:"name,omitempty"`
	Args []string `json:"args,omitempty"`
}
type Rule struct {
	Api     string
	Actions []*Action
}

type Action struct {
	Action       *FuncExpr
	ComputeKey   string
	Terminal     *bool
	Condition    *FuncExpr
	ConditionAnd []*FuncExpr
	ConditionOr  []*FuncExpr
}

type ActionResult struct{}

type IFunc interface {
	Name() string
	ValidateParamRef(paramRefs []ParamRef) error
	Call(ctx context.Context, cxCtx *ExecutionState, args []string) (string, error)
}

type ParamPathPrefix int

const (
	ParamPathPrefixInvalid ParamPathPrefix = iota
	ParamPathPrefixProto
	ParamPathPrefixLiteral
	ParamPathPrefixHeader
	ParamPathPrefixCompute
	ParamPathPrefixBody
	ParamPathPrefixRule
	ParamPathPrefixEnv
	ParamPathPrefixProcessEnv
	paramPathPrefixMax
)

func ToParamPathPrefix(prefix string) ParamPathPrefix {
	return prefixMap[prefix]
}

var prefixMap = map[string]ParamPathPrefix{}

func init() {
	for i := ParamPathPrefixInvalid + 1; i < paramPathPrefixMax; i++ {
		name := i.String()
		if name == "" {
			panic("invalid param path prefix")
		}
		prefixMap[name] = i
	}
}

func (prefix ParamPathPrefix) String() string {
	switch prefix {
	case ParamPathPrefixProto:
		return "proto"
	case ParamPathPrefixLiteral:
		return "literal"
	case ParamPathPrefixHeader:
		return "header"
	case ParamPathPrefixCompute:
		return "compute"
	case ParamPathPrefixBody:
		return "body"
	case ParamPathPrefixRule:
		return "rule"
	case ParamPathPrefixEnv:
		return "env"
	case ParamPathPrefixProcessEnv:
		return "process_env"
	default:
		return ""
	}
}

// ParamRef contain ref info the real param value.
type ParamRef struct {
	Prefix ParamPathPrefix
	Path   string
}

func ParseParamRef(ref string) (ParamRef, error) {
	parts := strings.SplitN(ref, ".", 2)
	if len(parts) < 1 {
		return ParamRef{}, fmt.Errorf("invalid param ref: %s", ref)
	}
	prefix := ToParamPathPrefix(parts[0])
	if prefix == ParamPathPrefixInvalid {
		return ParamRef{}, fmt.Errorf("invalid param ref: %s", ref)
	}
	var p string
	if len(parts) >= 2 {
		p = parts[1]
	}
	return ParamRef{
		Prefix: prefix,
		Path:   p,
	}, nil
}

type FuncExpr struct {
	Not       bool
	F         IFunc
	ParamRefs []ParamRef
}

type ProtocolSpec interface {
	Get(key string) string
}

type Header interface {
	Get(key string) string
}

type Body interface {
	Get(key string) string
}

type RouteType int

const (
	RouteTypeInvalid RouteType = iota
	RouteTypeAny
	RouteTypeFix
	RouteTypeTry
)
