package golangapirouter

import (
	"context"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

func (rule *Rule) init(ruleConf *RuleConfig) error {
	rule.Api = ruleConf.Api
	return nil
}

// Executor execute rule and fill result to ExecutionState
type Executor struct {
	funcs map[string]IFunc
}

func NewExecutor() *Executor {
	executor := &Executor{
		funcs: make(map[string]IFunc),
	}
	executor.MustRegisterFunc(builtinIn{})
	executor.MustRegisterFunc(builtinHasPrefix{})
	executor.MustRegisterFunc(builtinConsult{})
	executor.MustRegisterFunc(builtinRouteConsultAndRetry{})
	executor.MustRegisterFunc(builtinRouteAny{})
	return executor
}

func (exec *Executor) MustRegisterFunc(f IFunc) {
	ff := exec.funcs[f.Name()]
	if ff != nil {
		panic("duplicate function name: " + f.Name())
	}
	exec.funcs[f.Name()] = f
}

func (exec *Executor) ParseRuleConfig(ruleConf *RuleConfig) (*Rule, error) {
	var err error
	rule := &Rule{}
	rule.Api = ruleConf.Api
	rule.Actions = make([]*Action, len(ruleConf.Actions))
	for i, actionConf := range ruleConf.Actions {
		action := &Action{
			ComputeKey: actionConf.ComputeKey,
			Terminal:   actionConf.Terminal,
		}
		action.Action, err = exec.parseFuncCall(actionConf.Name, actionConf.Args)
		if err != nil {
			return nil, err
		}
		action.Condition, err = exec.parseFuncCall(actionConf.Condition, actionConf.ConditionArgs)
		if err != nil {
			return nil, err
		}
		action.ConditionAnd = make([]*FuncExpr, len(actionConf.ConditionAnd))
		for i, condition := range actionConf.ConditionAnd {
			action.ConditionAnd[i], err = exec.parseFuncCall(condition.Name, condition.Args)
			if err != nil {
				return nil, err
			}
		}
		action.ConditionOr = make([]*FuncExpr, len(action.ConditionOr))
		for i, condition := range actionConf.ConditionOr {
			action.ConditionOr[i], err = exec.parseFuncCall(condition.Name, condition.Args)
			if err != nil {
				return nil, err
			}
		}
		rule.Actions[i] = action
	}
	return rule, nil
}

func (exec *Executor) parseFuncCall(name string, paramRefs []string) (*FuncExpr, error) {
	if name == "" && len(paramRefs) == 0 {
		return nil, nil
	}
	expr := &FuncExpr{}
	if strings.HasSuffix(name, "!") {
		expr.Not = true
		name = name[:len(name)-1]
	}
	expr.F = exec.funcs[name]
	if expr.F == nil {
		return nil, fmt.Errorf("function %s not found", name)
	}
	for _, ref := range paramRefs {
		paramRef, err := ParseParamRef(ref)
		if err != nil {
			return nil, err
		}
		expr.ParamRefs = append(expr.ParamRefs, paramRef)
	}
	return expr, nil
}

func (exec *Executor) Execute(ctx context.Context, state *ExecutionState) error {
	for i, action := range state.rule.Actions {
		state.actionIndex = i
		ok, err := exec.evaluateConditions(ctx, state, action)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		params := state.ComputeParams(ctx, action.Action.ParamRefs)
		_, err = action.Action.F.Call(ctx, state, params)
		if err != nil {
			return err
		}
		if state.routeType != RouteTypeInvalid {
			break
		}
	}
	if state.routeType == RouteTypeInvalid {
		return ErrNoValidUnitDecided
	}
	return nil
}

func (exec *Executor) evaluateConditions(ctx context.Context, execCtx *ExecutionState, action *Action) (bool, error) {
	return true, nil
}

type ExecutionState struct {
	protoSpec   ProtocolSpec
	header      Header
	body        Body
	rule        *Rule
	actionIndex int
	compute     map[string]string
	routeType   RouteType
	unitIds     []int
}

func (exec *ExecutionState) SetRoutingResult(routeType RouteType, unitIds []int) {
	exec.routeType = routeType
	exec.unitIds = unitIds
}

func (exec *ExecutionState) ComputeParams(ctx context.Context, paramRefs []ParamRef) []string {
	args := make([]string, len(paramRefs))
	for i, paramRef := range paramRefs {
		switch paramRef.Prefix {
		case ParamPathPrefixCompute:
			args[i] = exec.getComputeValue(paramRef.Path)
		case ParamPathPrefixHeader:
			args[i] = exec.header.Get(paramRef.Path)
		case ParamPathPrefixLiteral:
			args[i] = paramRef.Path
		case ParamPathPrefixBody:
			args[i] = exec.body.Get(paramRef.Path)
		case ParamPathPrefixProto:
			args[i] = exec.protoSpec.Get(paramRef.Path)
		}

	}
	return args
}

func (exec *ExecutionState) getComputeValue(p string) string {
	dotIndex := strings.Index(p, ".")
	var restPath string
	if dotIndex == -1 {
		dotIndex = len(p)
	} else {
		restPath = p[dotIndex+1:]
	}
	val := exec.compute[p[:dotIndex]]
	result := gjson.Get(val, restPath)
	return result.String()
}
