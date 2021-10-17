package layout

type Graph struct {
	Edges map[[2]uint64]Edge
	Nodes map[uint64]Node
}

type Node struct {
	XY [2]int // smallest x,y corner
	W  int
	H  int
}

type Edge struct {
	Path [][2]int // [start: {x,y}, ... finish: {x,y}]
}

// utils, do not bind them as methods to emphasize that structs are pure data containers.

func CopyGraph(g Graph) Graph {
	ng := Graph{}
	for id, n := range g.Nodes {
		ng.Nodes[id] = n
	}
	for id, e := range g.Edges {
		nedges := make([][2]int, len(e.Path))
		copy(nedges, e.Path)
		ng.Edges[id] = Edge{
			Path: nedges,
		}
	}
	return ng
}

func roots(g Graph) []uint64 {
	var roots []uint64
	for n := range g.Nodes {
		hasParents := false
		for e := range g.Edges {
			if e[1] == n {
				hasParents = true
			}
		}
		if !hasParents {
			roots = append(roots, n)
		}
	}
	return roots
}

func totalNodesWidth(g Graph) int {
	w := 0
	for _, node := range g.Nodes {
		w += node.W
	}
	return w
}

func totalNodesHeight(g Graph) int {
	h := 0
	for _, node := range g.Nodes {
		h += node.H
	}
	return h
}
