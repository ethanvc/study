package golangapirouter

import "context"

type RuleConfig struct {
	Api     string          `json:"api"`
	Actions []*ActionConfig `json:"actions"`
}

type ActionConfig struct {
	Expr         string   `json:"expr,omitempty"`
	ExprArgs     []string `json:"expr_args,omitempty"`
	ExprAnd      []Expr   `json:"expr_and,omitempty"`
	ExprOr       []Expr   `json:"expr_or,omitempty"`
	Name         string   `json:"name"`
	Args         []string `json:"args,omitempty"`
	ComputeKey   string   `json:"compute_key,omitempty"`
	Terminal     *bool    `json:"terminal,omitempty"`
	PreferUnitId *int     `json:"prefer_unit_id,omitempty"`
}

type Expr struct {
	Expr     string   `json:"expr,omitempty"`
	ExprArgs []string `json:"expr_args,omitempty"`
}
type Rule struct {
	Api     string
	Actions []*Action
}

type Action struct {
	Condition *FuncExpr
	Action    FuncExpr
}

type ActionResult struct{}

type IFunc interface {
	Call(ctx context.Context, cxCtx *ExecutionState, args []string) (string, error)
}

type ParamPathPrefix int

const (
	ParamPathPrefixProto ParamPathPrefix = iota + 1
	ParamPathPrefixLiteral
	ParamPathPrefixHeader
	ParamPathPrefixCompute
	ParamPathPrefixBody
	ParamPathPrefixRule
	ParamPathPrefixEnv
	ParamPathPrefixProcessEnv
)

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
