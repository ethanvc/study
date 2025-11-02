package ginradix

import "testing"

func newTestTree() *Tree[int] {
	return &Tree[int]{}
}

func Test_Basic(t *testing.T) {
	tree := newTestTree()
	tree.MustInsert("/abc/:id", 1)
}
