package golangapirouter

import (
	"context"
	"github.com/ethanvc/study/golangproj/logjson/internal/json"
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
