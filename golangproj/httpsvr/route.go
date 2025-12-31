package httpsvr

import "github.com/ethanvc/study/golangproj/httpsvr/ginradix"

type Router struct {
	raidx ginradix.Tree[*RouteNode]
}

func (r *Router) Get(method string, pattern string) {
	// n, params := r.raidx.Search(pattern, nil)
}

type RouteNode struct {
	nodes []Node
}

type Node struct {
}
