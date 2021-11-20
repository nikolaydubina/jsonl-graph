package dot

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"github.com/nikolaydubina/jsonl-graph/graph"
)

type Colorer interface {
	Color(k string, v interface{}) color.Color
}

// ColoredNodeLabel is label content for colorized Graphviz node
type ColoredNodeLabel struct {
	n       graph.NodeData
	colorer Colorer
}

func (r ColoredNodeLabel) Render() string {
	rows := []string{}
	for k, v := range r.n {
		if k == "id" || strings.HasSuffix(k, "_url") {
			continue
		}

		row := fmt.Sprintf(`
			<tr>
				<td border="1" ALIGN="LEFT">%s</td>
				<td border="1" ALIGN="RIGHT" bgcolor="%s">%s</td>
			</tr>`,
			k,
			Color{c: r.colorer.Color(k, v)}.Render(),
			Value{v: v}.Render(),
		)

		rows = append(rows, row)
	}

	// this will sort by key, since key is first
	sort.Strings(rows)

	return strings.Join(
		[]string{
			"<<table border=\"0\" cellspacing=\"0\" CELLPADDING=\"6\">",
			fmt.Sprintf(`
				<tr>
					<td port="port0" border="1" colspan="2" ALIGN="CENTER" bgcolor="%s">%s</td>
				</tr>`,
				Color{c: color.RGBA{R: 200, G: 200, B: 200, A: 200}}.Render(),
				Value{v: r.n["id"]}.Render(),
			),
			strings.Join(rows, "\n"),
			"</table>>",
		},
		"\n",
	)
}

// Color transforms Go color to Graphviz RGBA format which is slightly odd
type Color struct {
	c color.Color
}

func (s Color) Render() string {
	r, g, b, a := s.c.RGBA()
	return fmt.Sprintf("#%x%x%x%x", uint8(r), uint8(g), uint8(b), uint8(a))
}

// NewColoredGraph creates renderable colored graph from graph data
func NewColoredGraph(
	graph graph.Graph,
	orientation Orientation,
	colorer Colorer,
) BasicGraph {
	nodes := make([]Renderable, 0, len(graph.Nodes))
	for _, n := range graph.Nodes {
		node := Node{id: n.ID(), shape: NoneShape, label: ColoredNodeLabel{n: n, colorer: colorer}}
		nodes = append(nodes, node)
	}

	edges := make([]Renderable, 0, len(graph.Edges))
	for _, e := range graph.Edges {
		edges = append(edges, BasicEdge{from: e.From(), to: e.To()})
	}

	return BasicGraph{
		orientation: orientation,
		nodes:       nodes,
		edges:       edges,
	}
}
