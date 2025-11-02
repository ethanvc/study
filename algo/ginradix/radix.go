package ginradix

import (
	"fmt"
	"strings"
)
import "errors"

type Tree[Value any] struct {
	root *Node[Value]
}

type Node[Value any] struct {
	Part      string
	Pattern   string
	Children  []*Node[Value]
	WildChild *Node[Value]
	Val       Value
	ValValid  bool
}

func newNode[Value any](part string) *Node[Value] {
	return &Node[Value]{
		Part: part,
	}
}

func (n *Node[Value]) insert(pattern, restPattern string, val Value) error {
	if n.isWildNode() {
		return n.insertInWildNode(pattern, restPattern, val)
	}
	if patternStartWithWild(restPattern) {
		if n.WildChild != nil {
			return n.WildChild.insert(pattern, restPattern, val)
		}
		var err error
		n.WildChild, err = createNewNodes(pattern, restPattern, val)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func patternStartWithWild(pattern string) bool {
	return pattern[0] == ':' || pattern[0] == '*'
}

func (n *Node[Value]) insertInWildNode(pattern, restPattern string, val Value) error {
	part := getNextPatternPart(restPattern)
	if part != n.Part {
		return fmt.Errorf("conflict param part(old=%s, new=%s), new pattern is %s", n.Part, part, pattern)
	}
	if len(part) == len(restPattern) {
		if n.ValValid {
			return fmt.Errorf("pattern %s already exist", pattern)
		}
		n.setVal(pattern, val)
		return nil
	}
	restPattern = restPattern[len(part):]
	candidate := n.getCandidate(restPattern)
	if candidate != nil {
		return candidate.insert(pattern, restPattern, val)
	}
	head, err := createNewNodes(pattern, restPattern, val)
	if err != nil {
		return err
	}
	n.insertChild(head)
	return nil
}

func (n *Node[Value]) getCandidate(pattern string) *Node[Value] {
	ch := pattern[0]
	if ch == ':' || ch == '*' {
		return n.WildChild
	}
	for _, child := range n.Children {
		if child.Part[0] == ch {
			return child
		}
	}
	return nil
}

func (n *Node[Value]) setVal(pattern string, val Value) {
	n.Pattern = pattern
	n.Val = val
	n.ValValid = true
}

func (n *Node[Value]) isWildNode() bool {
	return n.Part[0] == ':' || n.Part[0] == '*'
}

func (n *Node[Value]) insertChild(child *Node[Value]) {
	if child.isWildNode() {
		n.WildChild = child
	} else {
		n.Children = append(n.Children, child)
	}
}

func (t *Tree[Value]) MustInsert(pattern string, val Value) {
	err := t.Insert(pattern, val)
	if err != nil {
		panic(err)
	}
}
func (t *Tree[Value]) Insert(pattern string, val Value) error {
	if pattern == "" {
		return errors.New("empty pattern")
	}
	if pattern[0] != '/' {
		return errors.New("pattern must start with '/'")
	}
	var err error
	if t.root == nil {
		t.root, err = createNewNodes(pattern, pattern, val)
		return err
	}
	return t.root.insert(pattern, pattern, val)
}

func createNewNodes[Value any](pattern string, restPattern string, val Value) (*Node[Value], error) {
	var head *Node[Value]
	var tail *Node[Value]
	for restPattern != "" {
		part := getNextPatternPart(restPattern)
		tmp := newNode[Value](part)
		if head == nil {
			head = tmp
		}
		if tail != nil {
			tail.insertChild(tmp)
		}
		tail = tmp
		restPattern = restPattern[len(part):]
	}
	tail.setVal(pattern, val)
	return head, nil
}

func getNextPatternPart(pattern string) string {
	firstByte := pattern[0]
	if firstByte == ':' || firstByte == '*' {
		idx := strings.IndexByte(pattern, '/')
		if idx != -1 {
			return pattern[:idx]
		}
		return pattern
	}

	slashFound := false
	for i, ch := range pattern {
		if slashFound {
			if ch == ':' || ch == '*' {
				return pattern[:i]
			}
		}
		if ch == '/' {
			slashFound = true
		} else {
			slashFound = false
		}
	}
	return pattern
}
