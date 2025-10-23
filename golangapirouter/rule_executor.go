package golangapirouter

import (
	"context"
	"strings"

	"github.com/tidwall/gjson"
)

// RuleExecutor execute rule and fill result to RuleExecutionState
type RuleExecutor struct {
}

func NewRuleExecutor() *RuleExecutor {
	return &RuleExecutor{}
}

func (exec *RuleExecutor) Execute(ctx context.Context, state *RuleExecutionState) error {
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

func (exec *RuleExecutor) evaluateConditions(ctx context.Context, execCtx *RuleExecutionState, action *Action) (bool, error) {
	return true, nil
}

type RuleExecutionState struct {
	ProtocolSpec ProtocolSpec
	Header       Header
	Body         Body
	rule         *Rule
	actionIndex  int
	compute      map[string]string
	routeType    RouteType
	unitIds      []int
}

func (exec *RuleExecutionState) SetRoutingResult(routeType RouteType, unitIds []int) {
	exec.routeType = routeType
	exec.unitIds = unitIds
}

func (exec *RuleExecutionState) ComputeParams(ctx context.Context, paramRefs []ParamRef) []string {
	args := make([]string, len(paramRefs))
	for i, paramRef := range paramRefs {
		switch paramRef.Prefix {
		case ParamPathPrefixCompute:
			args[i] = exec.getComputeValue(paramRef.Path)
		case ParamPathPrefixHeader:
			args[i] = exec.Header.Get(paramRef.Path)
		case ParamPathPrefixLiteral:
			args[i] = paramRef.Path
		case ParamPathPrefixBody:
			args[i] = exec.Body.Get(paramRef.Path)
		case ParamPathPrefixProto:
			args[i] = exec.ProtocolSpec.Get(paramRef.Path)
		}

	}
	return args
}

func (exec *RuleExecutionState) getComputeValue(p string) string {
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
