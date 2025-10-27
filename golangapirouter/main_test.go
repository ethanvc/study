package golangapirouter

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var sRulesMap map[string]*RuleConfig = make(map[string]*RuleConfig)

func getRule(executor *Executor, api string) *Rule {
	ruleConf := sRulesMap[api]
	if ruleConf == nil {
		panic("rule conf not found: " + api)
	}
	rule, err := executor.ParseRuleConfig(ruleConf)
	if err != nil {
		panic(err)
	}
	return rule
}

func Test_Basic(t *testing.T) {
	ctx := context.Background()
	executor := NewExecutor()
	initExecutor(t, executor)
	req := httptest.NewRequest("GET", "/", nil)
	rule := getRule(executor, "/abc/fast")
	state := NewExecutionStateFromHttpRequest(rule, req)
	err := executor.Execute(ctx, state)
	require.NoError(t, err)
}

func init() {
	contents, err := os.ReadFile("rule.json")
	if err != nil {
		panic(err)
	}
	type RuleFile struct {
		Rules []*RuleConfig `json:"rules"`
	}
	var ruleFile RuleFile
	err = json.Unmarshal(contents, &ruleFile)
	if err != nil {
		panic(err)
	}
	for _, rule := range ruleFile.Rules {
		sRulesMap[rule.Api] = rule
	}
}

func initExecutor(t *testing.T, executor *Executor) {
	executor.MustRegisterFunc(testConsult{})
	executor.MustRegisterFunc(testRouteConsultAndRetry{})
	executor.MustRegisterFunc(testIn{})
}

type testConsult struct{}

func (testConsult) Name() string {
	return "_consult"
}

func (testConsult) ValidateParamRef(paramRefs []ParamRef) error {
	return nil
}

func (testConsult) Call(ctx context.Context, cxCtx *ExecutionState, args []string) (string, error) {
	return "", nil
}

type testRouteConsultAndRetry struct {
}

func (testRouteConsultAndRetry) Name() string {
	return "_route_consult_and_retry"
}
func (testRouteConsultAndRetry) ValidateParamRef(paramRefs []ParamRef) error {
	return nil
}
func (testRouteConsultAndRetry) Call(ctx context.Context, cxCtx *ExecutionState, args []string) (string, error) {
	return "", nil
}

type testIn struct {
}

func (testIn) Name() string {
	return "_in"
}

func (testIn) ValidateParamRef(paramRefs []ParamRef) error {
	return nil
}
func (testIn) Call(ctx context.Context, cxCtx *ExecutionState, args []string) (string, error) {
	return "", nil
}
