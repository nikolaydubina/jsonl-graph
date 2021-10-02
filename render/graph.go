package render

import (
	"fmt"
	"strings"
)

// Graph is rendered graph.
type Graph struct {
	Nodes map[uint64]*Node
	Edges map[uint64]map[uint64]*Edge
}

// NewGraph initializes empty Graph.
func NewGraph() Graph {
	return Graph{
		Nodes: map[uint64]*Node{},
		Edges: map[uint64]map[uint64]*Edge{},
	}
}

// Render creates root svg element
func (g Graph) RenderSVGRoot(rootID string) string {
	body := []string{
		fmt.Sprintf(`<g id="%s">`, rootID),
	}

	for _, tos := range g.Edges {
		for _, edge := range tos {
			body = append(body, edge.Render())
		}
	}

	// draw nodes always on top of edges
	for _, node := range g.Nodes {
		body = append(body, node.Render())
	}

	body = append(body, "</g>")

	return strings.Join(body, "\n")
}

// Render creates SVG.
func (g Graph) Render(svgID, rootID string) string {
	defs := []string{
		arrowDef(),
	}
	return svg(defs, svgID, g.RenderSVGRoot(rootID))
}

// TotalNodesWidth is sum of all nodes width.
func (g Graph) TotalNodesWidth() int {
	w := 0
	for _, node := range g.Nodes {
		w += node.Width()
	}
	return w
}

// TotalNodesHeight is sum of all nodes height.
func (g Graph) TotalNodesHeight() int {
	h := 0
	for _, node := range g.Nodes {
		h += node.Height()
	}
	return h
}

// BoundingBox coordinates that should fit whole graph.
func (g Graph) BoundingBox() (minx, miny, maxx, maxy int) {
	for _, node := range g.Nodes {
		nx := node.LeftBottom.X
		ny := node.LeftBottom.Y

		if nx < minx {
			minx = nx
		}
		if (nx + node.Width()) > maxx {
			maxx = nx + node.Width()
		}
		if ny < miny {
			miny = ny
		}
		if (ny + node.Height()) > maxy {
			maxy = ny + node.Height()
		}
	}
	return minx, miny, maxx, maxy
}

// Width returns expected highest X coordinate needed to render graph.
func (g Graph) Width() int {
	minx, _, maxx, _ := g.BoundingBox()
	return maxx - minx
}

// Height returns expected highest Y coordinate needed to render graph.
func (g Graph) Height() int {
	_, miny, _, maxy := g.BoundingBox()
	return maxy - miny
}

// Copy returns deep copy of current graph.
// TODO: make sure node and edges copied.
func (g Graph) Copy() Graph {
	other := NewGraph()
	for i := range g.Nodes {
		node := *g.Nodes[i]
		other.Nodes[i] = &node
	}
	for from, tos := range g.Edges {
		if _, ok := other.Edges[from]; !ok {
			other.Edges[from] = make(map[uint64]*Edge, len(tos))
		}
		for to := range tos {
			edge := *g.Edges[from][to]
			other.Edges[from][to] = &edge
		}
	}
	return other
}
