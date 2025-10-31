package radixtree

import (
	"bytes"
	"fmt"
)

type RadixTree[Processor PatternProcessor, Value any] struct {
	processor Processor
	node      *RadixNode[Value]
}

func NewRadixTree[Processor PatternProcessor, Value any]() *RadixTree[Processor, Value] {
	return &RadixTree[Processor, Value]{}
}

type RadixNode[Value any] struct {
	children    map[byte]*RadixNode[Value]
	paramChild  *RadixNode[Value]
	patternNode PatternNode
	val         Value
	valEnabled  bool
	pattern     string
}

func (n *RadixNode[Value]) reset() {
	n.children = nil
	n.paramChild = nil
	n.patternNode = PatternNode{}
	n.valEnabled = false
	var defaultVal Value
	n.val = defaultVal
	n.pattern = ""
}

func (n *RadixNode[Value]) isParamNode() bool {
	return n.patternNode.ParamType
}

func (n *RadixNode[Value]) insertChild(child *RadixNode[Value]) {
	if child.patternNode.ParamType {
		n.paramChild = child
	} else {
		if n.children == nil {
			n.children = make(map[byte]*RadixNode[Value])
		}
		n.children[getFirstByte(child.patternNode.NodeVal)] = child
	}
}

func (n *RadixNode[Value]) getCandidateChild(patternNode PatternNode) *RadixNode[Value] {
	if patternNode.ParamType {
		return n.paramChild
	} else {
		return n.children[getFirstByte(patternNode.NodeVal)]
	}
}

func (n *RadixNode[Value]) GetValue() Value {
	return n.val
}

type PatternNode struct {
	ParamType bool
	NodeVal   string
}

type Param struct {
	Key   string
	Value string
}

type Params []Param

type PatternProcessor interface {
	SplitPattern(p string) ([]PatternNode, error)
	GetParam(node PatternNode, p string) Param
}

func (t *RadixTree[Processor, Value]) Insert(pattern string, val Value) error {
	patternNodes, err := t.processor.SplitPattern(pattern)
	if err != nil {
		return err
	}
	n, err := t.insert(patternNodes)
	if err != nil {
		return err
	}
	n.pattern = pattern
	n.val = val
	n.valEnabled = true
	return nil
}

func (t *RadixTree[Processor, Value]) insert(newNodes []PatternNode) (*RadixNode[Value], error) {
	if t.node == nil {
		head, last := patternNodesToRadixNodes[Value](newNodes)
		t.node = head
		return last, nil
	}
	n := t.node
	var i int
	var newPatternNode PatternNode
	for i, newPatternNode = range newNodes {
		if n.patternNode != newPatternNode {
			break
		}
		if i == len(newNodes)-1 {
			if n.valEnabled {
				return nil, fmt.Errorf("pattern already exist: %s", getPattern(newNodes))
			}
			return n, nil
		}
		nextNewNode := newNodes[i+1]
		candidateNode := n.getCandidateChild(nextNewNode)
		if candidateNode == nil {
			head, last := patternNodesToRadixNodes[Value](newNodes)
			n.insertChild(head)
			return last, nil
		}
		n = candidateNode
	}
	if n.isParamNode() {
		if newPatternNode.ParamType {
			return nil, fmt.Errorf(" param node have different placeholder(old=%s, new=%s)",
				n.patternNode.NodeVal, newPatternNode.NodeVal)
		} else {

		}
	}
	if n.isParamNode() {
		// parent node must be a plain node
		if newPatternNode.ParamType {
			return nil, fmt.Errorf(" param node have different placeholder(old=%s, new=%s)",
				n.patternNode.NodeVal, newPatternNode.NodeVal)
		} else {
			panic("parent node is plain node, child node will never be plain node")
		}
	}
	if newPatternNode.ParamType {
		panic("param node will never be plain node here")
	}
	prefix := longestCommonPrefix(n.patternNode.NodeVal, newPatternNode.NodeVal)
	if n.patternNode.NodeVal == prefix {
		newNodes[i].NodeVal = newPatternNode.NodeVal[len(prefix):]
		head, tail := patternNodesToRadixNodes[Value](newNodes)
		n.insertChild(head)
		return tail, nil
	}
	if prefix == newPatternNode.NodeVal {
		oldChild := *n
		oldChild.patternNode.NodeVal = newPatternNode.NodeVal[len(prefix):]
		n.reset()
		n.insertChild(&oldChild)
		return n, nil
	}
	oldChild := *n
	oldChild.patternNode.NodeVal = n.patternNode.NodeVal[len(prefix):]
	n.reset()
	n.patternNode.NodeVal = prefix
	n.insertChild(&oldChild)
	newNodes[i].NodeVal = newPatternNode.NodeVal[len(prefix):]
	head, tail := patternNodesToRadixNodes[Value](newNodes)
	n.insertChild(head)
	return tail, nil
}

func (t *RadixTree[Processor, Value]) MustInsert(pattern string, val Value) {
	err := t.Insert(pattern, val)
	if err != nil {
		panic(err)
	}
}

func getPattern(patternNodes []PatternNode) string {
	buf := bytes.NewBuffer(nil)
	for _, patternNode := range patternNodes {
		buf.WriteString(patternNode.NodeVal)
	}
	return buf.String()
}

func patternNodesToRadixNodes[Value any](patternNodes []PatternNode) (head, last *RadixNode[Value]) {
	for _, patternNode := range patternNodes {
		if head == nil {
			head = &RadixNode[Value]{
				patternNode: patternNode,
			}
			last = head
			continue
		}
		if patternNode.ParamType {
			last.paramChild = &RadixNode[Value]{
				patternNode: patternNode,
			}
			last = last.paramChild
		} else {
			last.children = make(map[byte]*RadixNode[Value])
			newChild := &RadixNode[Value]{
				patternNode: patternNode,
			}
			last.children[patternNode.NodeVal[0]] = newChild
			last = newChild
		}
	}
	return
}

func (t *RadixTree[Processor, Value]) Search(p string, params Params) (*RadixNode[Value], Params, error) {
	n := t.node
	var traceInfo traceBackInfo[Value]
	traceInfo.OriginPath = p
	traceInfo.Path = &p
	traceInfo.Params = params
	for {
		if n == nil {
			if n = traceInfo.Pop(); n == nil {
				return nil, params, nil
			}
			goto BackTracePoint
		}
		if n.patternNode.ParamType {
			param := t.processor.GetParam(n.patternNode, p)
			traceInfo.Params = append(traceInfo.Params, param)
			p = p[len(param.Value):]
		} else {
			nodeLen := len(n.patternNode.NodeVal)
			if nodeLen > len(p) {
				n = traceInfo.Pop()
				if n == nil {
					return nil, params, nil
				}
				goto BackTracePoint
			}
			if n.patternNode.NodeVal != p[0:nodeLen] {
				n = traceInfo.Pop()
				if n == nil {
					return nil, params, nil
				}
				goto BackTracePoint
			}
			p = p[nodeLen:]
		}
		if p == "" {
			if n.valEnabled {
				return n, traceInfo.Params, nil
			} else {
				n = traceInfo.Pop()
				if n == nil {
					return nil, params, nil
				}
				goto BackTracePoint
			}
		}
		traceInfo.Push(p, n)
		n = n.children[p[0]]
		continue

	BackTracePoint:
		n = n.paramChild
	}
}

func getFirstByte(s string) byte {
	if s == "" {
		return 0
	}
	return s[0]
}

func longestCommonPrefix(s1, s2 string) string {
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}

	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			return s1[:i] // 返回从0到i-1的子串
		}
	}

	return s1[:minLen]
}

type backTraceNode[Value any] struct {
	ParamSize       int
	TargetPathStart int
	Node            *RadixNode[Value]
}

type backTraceNodes[Value any] []backTraceNode[Value]

func (t backTraceNodes[Value]) Push(n *RadixNode[Value], size int) backTraceNodes[Value] {
	return append(t, backTraceNode[Value]{
		ParamSize: size,
		Node:      n,
	})
}

func (t backTraceNodes[Value]) Pop(params Params) (backTraceNodes[Value], *RadixNode[Value], Params) {
	if len(t) == 0 {
		return nil, nil, params
	}
	lastNode := t[len(t)-1]
	return t[0 : len(t)-1], lastNode.Node, params[0:lastNode.ParamSize]
}

type traceBackInfo[Value any] struct {
	MatchParamNode bool
	OriginPath     string
	Params         Params
	TraceNodes     backTraceNodes[Value]
	Path           *string
}

func (info *traceBackInfo[Value]) Push(p string, n *RadixNode[Value]) {
	info.TraceNodes = append(info.TraceNodes, backTraceNode[Value]{
		ParamSize:       len(info.Params),
		TargetPathStart: len(info.OriginPath) - len(p),
		Node:            n,
	})
}

func (info *traceBackInfo[Value]) Pop() *RadixNode[Value] {
	if len(info.TraceNodes) == 0 {
		return nil
	}
	info.MatchParamNode = true
	last := info.TraceNodes[len(info.TraceNodes)-1]
	info.TraceNodes = info.TraceNodes[0 : len(info.TraceNodes)-1]

	info.Params = info.Params[0:last.ParamSize]
	*info.Path = info.OriginPath[last.TargetPathStart:]
	return last.Node
}
