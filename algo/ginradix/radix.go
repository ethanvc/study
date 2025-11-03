package ginradix

import (
	"errors"
	"fmt"
	"strings"
)

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

func (n *Node[Value]) reset(part string) {
	n.Part = part
	n.Pattern = ""
	n.Children = nil
	n.WildChild = nil
	var defaultVal Value
	n.Val = defaultVal
	n.ValValid = false
}

func (n *Node[Value]) insert(pattern, restPattern string, val Value) error {
	if n.isWildNode() {
		return n.insertWild(pattern, restPattern, val)
	}
	return n.insertPlain(pattern, restPattern, val)
}

// n is a plain node
func (n *Node[Value]) insertPlain(pattern, restPattern string, val Value) error {
	var err error
	if patternStartWithWild(restPattern) {
		// a vs :id
		return fmt.Errorf("ShouldNeverComeHere; pattern=%s, n.part=%s", pattern, n.Part)
	}
	currentPart := getNextPatternPart(restPattern)
	prefix := longestCommonPrefix(n.Part, currentPart)
	prefixLen := len(prefix)
	restPattern = restPattern[prefixLen:]
	if prefixLen == len(currentPart) && prefixLen == len(n.Part) {
		// a vs a
		if restPattern == "" {
			if n.ValValid {
				return fmt.Errorf("PatternAlreadyExist;pattern=%s", pattern)
			}
			n.setVal(pattern, val)
			return nil
		}
		candidate := n.getCandidate(restPattern)
		if candidate == nil {
			head, err := createNewNodes(pattern, restPattern, val)
			if err != nil {
				return err
			}
			n.insertChild(head)
			return nil
		}
		return candidate.insert(pattern, restPattern, val)
	}
	if prefixLen == len(n.Part) {
		// a vs ab
		candidate := n.getCandidate(restPattern)
		if candidate != nil {
			return candidate.insertPlain(pattern, restPattern, val)
		}
		newChild := newNode[Value](currentPart[prefixLen:])
		err = newChild.insert(pattern, restPattern, val)
		if err != nil {
			return err
		}
		n.insertChild(newChild)
		return nil
	}
	if prefixLen == len(prefix) {
		// ab vs a
		newChild, err := createNewNodes(pattern, restPattern, val)
		if err != nil {
			return err
		}
		oldChild := *n
		n.reset(prefix)
		n.setVal(pattern, val)
		oldChild.Part = currentPart[prefixLen:]
		n.insertChild(&oldChild)
		n.insertChild(newChild)
		return nil
	}
	// ab vs ac
	newChild := newNode[Value](currentPart[prefixLen:])
	if restPattern == "" {
		newChild.setVal(pattern, val)
	} else {
		restChildren, err := createNewNodes(pattern, restPattern, val)
		if err != nil {
			return err
		}
		newChild.insertChild(restChildren)
	}
	oldChild := *n
	oldChild.Part = oldChild.Part[prefixLen:]
	n.reset(prefix)
	n.insertChild(&oldChild)
	n.insertChild(newChild)
	return nil
}

func patternStartWithWild(pattern string) bool {
	return pattern[0] == ':' || pattern[0] == '*'
}

// n is a wild node
func (n *Node[Value]) insertWild(pattern, restPattern string, val Value) error {
	part := getNextPatternPart(restPattern)
	if !patternStartWithWild(part) {
		// :id vs a
		return fmt.Errorf("ShouldNeverHere: pattern=%s, n.part=%s", pattern, n.Part)
	}
	if part != n.Part {
		// :id vs :idx
		return fmt.Errorf("PatternConflict: pattern=%s, n.part=%s", pattern, n.Part)
	}
	// :id vs :id
	restPattern = restPattern[len(part):]
	if restPattern == "" {
		if n.ValValid {
			return fmt.Errorf("PatternAlreadyExist; pattern=%s", pattern)
		}
		n.setVal(pattern, val)
		return nil
	}
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
