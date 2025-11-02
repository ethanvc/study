package prefixmatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Tree struct {
	children  map[byte]*Node
	emptyNode *Node
}

type Node struct {
	Part       string
	Pattern    string
	Children   map[byte]*Node
	Value      int
	ValueValid bool
}

func newNode(pattern, partPattern string, value int) *Node {
	return &Node{
		Part:       partPattern,
		Pattern:    pattern,
		Value:      value,
		ValueValid: true,
	}
}

func (n *Node) insertChild(child *Node) {
	if n.Children == nil {
		n.Children = make(map[byte]*Node)
	}
	n.Children[child.Part[0]] = child
}

func (n *Node) reset() {
	n.Part = ""
	n.Pattern = ""
	n.Children = make(map[byte]*Node)
	n.Value = 0
	n.ValueValid = false
}

func (n *Node) candidateChild(pattern string) *Node {
	return n.Children[pattern[0]]
}

func (n *Node) insert(pattern string, part string, value int) error {
	prefix := longestCommonPrefix(n.Part, part)
	prefixLen := len(prefix)
	// use len to check eq, have better performance
	if prefixLen == len(part) && prefixLen == len(n.Part) {
		// a vs a
		if n.ValueValid {
			return fmt.Errorf("%s already exist", pattern)
		}
		n.ValueValid = true
		n.Value = value
		return nil
	}
	if prefixLen == len(n.Part) {
		// a vs ab
		newPart := part[prefixLen:]
		candidate := n.candidateChild(newPart)
		if candidate == nil {
			newChild := newNode(pattern, part[prefixLen:], value)
			n.insertChild(newChild)
			return nil
		}
		return candidate.insert(pattern, newPart, value)
	}
	if prefixLen == len(part) {
		// ab vs a
		oldChild := *n
		oldChild.Part = oldChild.Part[prefixLen:]
		n.reset()
		n.insertChild(&oldChild)
		n.Value = value
		n.ValueValid = true
		n.Part = prefix
		return nil
	}
	// ab vs ac
	oldChild := *n
	oldChild.Part = oldChild.Part[prefixLen:]
	n.reset()
	n.Part = prefix
	n.Children[oldChild.Part[0]] = &oldChild
	newChild := newNode(pattern, part[prefixLen:], value)
	n.Children[newChild.Part[0]] = newChild
	return nil
}

// NewTree 创建一个新的基数树
func NewTree() *Tree {
	return &Tree{
		children: make(map[byte]*Node),
	}
}

func (t *Tree) debugPrint() string {
	panic("why here")
	buf, err := json.Marshal(t.children)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf))
	return string(buf)
}

func (t *Tree) Search(pattern string) *Node {
	return t.search(pattern)
}

// return the most matched node and rest Pattern
func (t *Tree) search(pattern string) *Node {
	if pattern == "" {
		return t.emptyNode
	}
	n := t.children[pattern[0]]
	if n == nil {
		return nil
	}
	part := pattern
	for {
		if !strings.HasPrefix(part, n.Part) {
			return nil
		}
		if len(n.Part) == len(part) {
			if n.ValueValid {
				return n
			} else {
				return nil
			}
		}
		newPart := part[len(n.Part):]
		candidate := n.candidateChild(newPart)
		if candidate == nil {
			return nil
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
		return errors.New("empty Pattern already in tree")
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

type NodeChildren struct {
	Nodes []*Node
}

func (c *NodeChildren) insert(n *Node) {
	existNode := c.getCandidateChild(n.Part)
	if existNode != nil {
		panic("duplicate node")
	}
	c.Nodes = append(c.Nodes, n)
}

func (c *NodeChildren) getCandidateChild(pattern string) *Node {
	patternCh := pattern[0]
	for _, n := range c.Nodes {
		if n.Part[0] == patternCh {
			return n
		}
	}
	return nil
}

type NodeChildren2 struct {
	FirstBytes string
	Nodes      []*Node
}

func (c *NodeChildren2) insert(n *Node) {
	existNode := c.getCandidateChild(n.Part)
	if existNode != nil {
		panic("duplicate node")
	}
	c.FirstBytes += n.Part[0:1]
	c.Nodes = append(c.Nodes, n)
}

func (c *NodeChildren2) getCandidateChild(pattern string) *Node {
	patternCh := pattern[0]
	for i, ch := range c.FirstBytes {
		if patternCh == byte(ch) {
			return c.Nodes[i]
		}
	}
	return nil
}
