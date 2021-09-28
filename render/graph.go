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
