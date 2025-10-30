package radixtree

import (
	"errors"
	"strings"
)

type PlainProcessor struct{}

func (p PlainProcessor) SplitPattern(pattern string) ([]PatternNode, error) {
	return []PatternNode{
		PatternNode{
			NodeVal: pattern,
		},
	}, nil
}

func (p PlainProcessor) GetParam(node PatternNode, pattern string) Param {
	panic("never come here")
}

func NewGinRadixTree[Value any]() *RadixTree[GinPatternProcessor, Value] {
	return NewRadixTree[GinPatternProcessor, Value]()
}

type GinPatternProcessor struct{}

func (p GinPatternProcessor) SplitPattern(pattern string) ([]PatternNode, error) {
	var nodes []PatternNode
	start := 0
	inParam := false
	lastChar := byte(0)
	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		if i != 0 {
			lastChar = pattern[i-1]
		}
		if inParam {
			if ch == '/' {
				inParam = false
				nodes = append(nodes, PatternNode{
					ParamType: true,
					NodeVal:   pattern[start:i],
				})
				start = i
			}
		} else {
			if (ch == ':' || ch == '*') && lastChar == '/' {
				nodes = append(nodes, PatternNode{NodeVal: pattern[start:i]})
				inParam = true
				start = i
			}
		}
	}
	if inParam {
		nodes = append(nodes, PatternNode{
			ParamType: true,
			NodeVal:   pattern[start:],
		})
	} else {
		nodes = append(nodes, PatternNode{
			NodeVal: pattern[start:],
		})
	}

	return nodes, nil
}

func (p GinPatternProcessor) GetParam(node PatternNode, path string) Param {
	param := Param{
		Key: node.NodeVal[1:],
	}
	if node.NodeVal[0] == '*' {
		param.Value = path
		return param
	}
	idx := strings.IndexByte(path, '/')
	if idx == -1 {
		param.Value = path
		return param
	}
	param.Value = path[:idx]
	return param
}

func getParamLen(pattern string, idx int) (int, error) {
	i := idx
	for ; i < len(pattern); i++ {
		ch := pattern[i]
		if i == idx {
			if !isChar(ch) {
				return 0, errors.New("param name must start with a letter")
			}
			continue
		}
		if ch == '/' {
			break
		}
		if isChar(ch) || isDigit(ch) {
			continue
		}
		return 0, errors.New("param name must only contain number and characters")
	}
	return i - idx, nil
}

func isChar(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func getPlainNodeLen(pattern string, idx int) (int, error) {
	i := idx
	for ; i < len(pattern); i++ {
		if pattern[i] != '/' {
			continue
		} else {
			break
		}
	}
	return i - idx, nil
}
