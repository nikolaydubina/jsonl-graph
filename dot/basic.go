package dot

import (
	"fmt"

	"github.com/nikolaydubina/jsonl-graph/graph"
)

type Renderable interface {
	Render() string
}

// Orientation should match Graphviz allowed values
type Orientation string

const (
	LR Orientation = "LR"
	TB Orientation = "TB"
)

// BasicGraph renders graph to Dot without colors with simples syntax without HTML.
// TODO: consider adding colors in background https://stackoverflow.com/questions/17765301/-dot-how-to-change-the-colour-of-one-record-in-multi-record-shape
type BasicGraph struct {
	orientation Orientation
	nodes       []Renderable
	edges       []Renderable
}

// NewBasicGraph creates renderable graph from graph data
func NewBasicGraph(
	graph graph.Graph,
	orientation Orientation,
) BasicGraph {
	nodes := make([]Renderable, 0, len(graph.Nodes))
	for _, n := range graph.Nodes {
		node := Node{id: n.ID(), shape: RecordShape, label: BasicNodeLabel{n: n}}
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

func (r BasicGraph) Render() string {
	s := "digraph G {\n"
	s += "rankdir=" + string(r.orientation) + "\n"

	for _, n := range r.nodes {
		s += n.Render() + "\n"

	}

	for _, e := range r.edges {
		s += e.Render() + "\n"
	}

	s += "}\n"

	return s
}

type BasicEdge struct {
	from string
	to   string
}

func (r BasicEdge) Render() string {
	return fmt.Sprintf(`"%s" -> "%s"`, r.from, r.to)
}
