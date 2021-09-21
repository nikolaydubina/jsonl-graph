package main

import (
	"log"
	"strings"
	"syscall/js"

	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/render"
	"github.com/nikolaydubina/jsonl-graph/web/svgpanzoom"
)

type layoutUpdater interface {
	UpdateGraphLayout(g render.Graph)
}

type Renderer struct {
	graphData     graph.Graph   // what graph contains
	graphRender   render.Graph  // how graph is rendered
	layoutUpdater layoutUpdater // how to make render graph
}

func (r Renderer) OnDataChange(_ js.Value, _ []js.Value) interface{} {
	inputString := js.Global().Get("document").Call("getElementById", "inputData").Get("value")

	g, err := graph.NewGraphFromJSONLReader(strings.NewReader(inputString.String()))
	if err != nil {
		log.Printf("bad input: %s", err)
		return nil
	}

	r.graphData.ReplaceFrom(g)
	r.Render()

	return nil
}

// UpdateRenderGraphWithDataGraph is called when graph data changed
// and we need to update render graph nodes and fields based on new
// data from data graph.
// We have to preserve ids and existing render information.
// For example, preserving positions in Nodes and Paths points in Edges.
func (r Renderer) UpdateRenderGraphWithDataGraph() {
	// update nodes with new data, preserve rest. add new nodes.
	for id, node := range r.graphData.Nodes {
		if _, ok := r.graphRender.Nodes[id]; !ok {
			r.graphRender.Nodes[id] = render.Node{}
		}

		rnode := r.graphRender.Nodes[id]
		rnode.Title = node.ID()

		r.graphRender.Nodes[id] = rnode
	}

	// delete render graph nodes that no longer present
	for id := range r.graphRender.Nodes {
		if _, ok := r.graphData.Nodes[id]; !ok {
			delete(r.graphRender.Nodes, id)
		}
	}

	// update edges with new data, preserve rest. add new edges.
	for fromID, edges := range r.graphData.Edges {
		if _, ok := r.graphRender.Edges[fromID]; !ok {
			r.graphRender.Edges[fromID] = make(map[uint64]render.Edge, len(edges))
		}

		// check all new data edges
		for toID := range edges {
			// new edge, creating new edge
			if _, ok := r.graphRender.Edges[fromID][toID]; !ok {
				r.graphRender.Edges[fromID][toID] = render.Edge{}
			}
			// existing edge. skipping, no fields to update.
		}
	}

	// delete render graph edges that no longer present
	for idFrom, edges := range r.graphRender.Edges {
		dEdges, ok := r.graphData.Edges[idFrom]
		if !ok {
			delete(r.graphRender.Edges, idFrom)
			continue
		}
		for idTo := range edges {
			if _, ok := dEdges[idTo]; !ok {
				delete(r.graphRender.Edges[idFrom], idTo)
			}
		}
	}
}

func (r Renderer) Render() {
	r.UpdateRenderGraphWithDataGraph()
	r.layoutUpdater.UpdateGraphLayout(r.graphRender)

	js.Global().
		Get("document").
		Call("getElementById", "output-container").
		Set("innerHTML", r.graphRender.Render())
}

func main() {
	c := make(chan bool)

	renderer := Renderer{
		graphRender: render.NewGraph(),
		graphData:   graph.NewGraph(),
		layoutUpdater: render.BasicGridLayout{
			W:         100,
			H:         16,
			RowLength: 10,
			Margin:    5,
		},
	}

	js.Global().
		Get("document").
		Call("getElementById", "inputData").
		Set("onkeyup", js.FuncOf(renderer.OnDataChange))

	renderer.OnDataChange(js.Value{}, nil)

	// once it is rendered at least once, bind handlers

	p := svgpanzoom.NewPanZoomer(
		"graph",
		0.2,
	)
	p.SetupHandlers()

	<-c
}
