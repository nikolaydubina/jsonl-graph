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
	ng := Graph{
		Nodes: make(map[uint64]Node, len(g.Nodes)),
		Edges: make(map[[2]uint64]Edge, len(g.Edges)),
	}
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
	hasParent := make(map[uint64]bool, len(g.Nodes))
	for e := range g.Edges {
		hasParent[e[1]] = true
	}

	var roots []uint64
	for n := range g.Nodes {
		if !hasParent[n] {
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

// BoundingBox coordinates that should fit whole graph.
func BoundingBox(g Graph) (minx, miny, maxx, maxy int) {
	for _, node := range g.Nodes {
		nx := node.XY[0]
		ny := node.XY[1]

		if nx < minx {
			minx = nx
		}
		if x := nx + node.W; x > maxx {
			maxx = x
		}
		if ny < miny {
			miny = ny
		}
		if y := ny + node.H; y > maxy {
			maxy = y
		}
	}
	return minx, miny, maxx, maxy
}
