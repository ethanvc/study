package prefixmatch

import (
	"fmt"
	"strings"
)

type Tree struct {
	root *Node
}

type Node struct {
	part       string
	pattern    string
	children   map[byte]*Node
	value      int
	valueValid bool
}

func (n *Node) insertChild(fullPattern, partPattern string, value int) {
	newChild := &Node{
		part:       partPattern,
		pattern:    fullPattern,
		value:      value,
		valueValid: true,
	}
	ch := partPattern[0]
	if n.children == nil {
		n.children = make(map[byte]*Node)
	}
	n.children[ch] = newChild
}

func (t *Tree) MustInsert(pattern string, value int) {
	err := t.Insert(pattern, value)
	if err != nil {
		panic(err)
	}
}

func (t *Tree) Insert(pattern string, value int) error {
	n, restPattern := t.search(pattern)
	if restPattern == "" {
		return fmt.Errorf("pattern %s already exist in tree", pattern)
	}
	if n == nil {
		t.root = &Node{
			part:       restPattern,
			pattern:    pattern,
			value:      value,
			valueValid: true,
		}
		return nil
	}
	n.insertChild(pattern, restPattern, value)
	return nil
}

func (t *Tree) Search(pattern string) *Node {
	n, restPattern := t.search(pattern)
	if restPattern != "" {
		return n
	}
	return nil
}

func (t *Tree) search(pattern string) (*Node, string) {
	n := t.root
	if n == nil {
		return nil, pattern
	}
	if !strings.HasPrefix(pattern, n.part) {
		return nil, pattern
	}
	for n != nil {

	}
	return n, pattern
}
