package render

import (
	"fmt"
	"image"
	"strings"
)

// Node is rendered point
type Node struct {
	LeftBottom image.Point
	Width      int
	Height     int
	Title      string
}

// TODO: for text https://developer.mozilla.org/en-US/docs/Web/SVG/Element/foreignObject
func (n Node) Render() string {
	return fmt.Sprintf(`
		<g>
			<rect x="%d" y="%d" width="%d" height="%d" style="fill:rgb(240,240,240);stroke-width:1;stroke:rgb(200,200,200);"></rect>
			<foreignObject x="%d" y="%d" width="%d" height="%d">
				<div xmlns="http://www.w3.org/1999/xhtml" style="font-size: 9px; overflow: hidden; text-align: center;">
				%s
				</div>
			</foreignObject>
		</g>
		`,
		n.LeftBottom.X,
		n.LeftBottom.Y,
		n.Width,
		n.Height,
		n.LeftBottom.X,
		n.LeftBottom.Y,
		n.Width,
		n.Height,
		n.Title,
	)
}

// Edge is rendered edge
type Edge struct {
	Points []image.Point
}

func (e Edge) Render() string {
	var points []string
	for _, point := range e.Points {
		points = append(points, fmt.Sprintf("%d,%d", point.X, point.Y))
	}
	return fmt.Sprintf(`<polyline style="fill:none;stroke-width:1;stroke: black;" points="%s"></polyline>`, strings.Join(points, " "))
}

// Graph is rendered graph.
type Graph struct {
	Nodes map[uint64]Node
	Edges map[uint64]map[uint64]Edge
}

// NewGraph initializes empty Graph.
func NewGraph() Graph {
	return Graph{
		Nodes: map[uint64]Node{},
		Edges: map[uint64]map[uint64]Edge{},
	}
}

// Render creates SVG.
func (g Graph) Render() string {
	defs := []string{
		arrowDef(),
	}

	body := []string{
		`<g id="graph">`,
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

	return svg(defs, body)
}
