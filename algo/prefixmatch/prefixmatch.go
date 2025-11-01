package prefixmatch

import (
	"errors"
	"fmt"
	"strings"
)

type Tree struct {
	children  map[byte]*Node
	emptyNode *Node
}

type Node struct {
	part       string
	pattern    string
	children   map[byte]*Node
	value      int
	valueValid bool
}

func newNode(pattern, partPattern string, value int) *Node {
	return &Node{
		part:       partPattern,
		pattern:    pattern,
		value:      value,
		valueValid: true,
	}
}

func (n *Node) reset() {
	n.part = ""
	n.pattern = ""
	n.children = make(map[byte]*Node)
	n.value = 0
	n.valueValid = false
}

func (n *Node) candidateChild(pattern string) *Node {
	return n.children[pattern[0]]
}

func (n *Node) insert(pattern string, part string, value int) error {
	prefix := longestCommonPrefix(n.part, part)
	if len(prefix) == len(n.part) {
		if n.valueValid {
			return fmt.Errorf("%s already exist", pattern)
		}
		n.valueValid = true
		n.value = value
		return nil
	}
	newPart := part[len(prefix):]
	if len(n.part) < len(prefix) {
		// existing pattern short then new pattern
		candidate := n.candidateChild(newPart)
		if candidate == nil {
			childNode := newNode(pattern, part[len(prefix):], value)
			n.children[childNode.part[0]] = childNode
			return nil
		}
		return candidate.insert(pattern, newPart, value)
	}
	// existing pattern longer then new pattern, split current node
	oldChild := *n
	oldChild.part = oldChild.part[len(prefix):]
	n.reset()
	n.part = prefix
	n.children[oldChild.part[0]] = &oldChild
	newChild := newNode(pattern, newPart, value)
	n.children[newChild.part[0]] = newChild
	return nil
}

// NewTree 创建一个新的基数树
func NewTree() *Tree {
	return &Tree{
		children: make(map[byte]*Node),
	}
}

func (t *Tree) Search(pattern string) *Node {
	n, restPattern := t.search(pattern)
	if n != nil && restPattern == "" && n.valueValid {
		return n
	} else {
		return nil
	}
}

// return the most matched node and rest pattern
func (t *Tree) search(pattern string) (*Node, string) {
	if pattern == "" {
		return t.emptyNode, pattern
	}
	n := t.children[pattern[0]]
	if n == nil {
		return nil, pattern
	}
	part := pattern
	for {
		if !strings.HasPrefix(part, n.part) {
			return n, part
		}
		if len(n.part) == len(part) {
			return n, ""
		}
		newPart := part[len(n.part):]
		candidate := n.candidateChild(newPart)
		if candidate == nil {
			return n, part
		}
		n = candidate
		part = newPart
	}
}

func (t *Tree) Insert(pattern string, value int) error {
	if pattern == "" {
		return t.insertEmptyPattern(value)
	}
	n := t.children[pattern[0]]
	if n == nil {
		t.children[pattern[0]] = newNode(pattern, pattern, value)
		return nil
	}
	return n.insert(pattern, pattern, value)
}

func (t *Tree) insertEmptyPattern(val int) error {
	if t.emptyNode == nil {
		t.emptyNode = newNode("", "", val)
		return nil
	} else {
		return errors.New("empty pattern already in tree")
	}
}

func longestCommonPrefix(a, b string) string {
	l := min(len(a), len(b))
	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			return a[0:i]
		}
	}
	return a[0:l]
}
