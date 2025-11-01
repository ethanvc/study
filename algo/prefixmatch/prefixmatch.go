package prefixmatch

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
