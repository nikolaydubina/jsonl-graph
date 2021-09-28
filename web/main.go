package main

import (
	"bytes"
	"log"
	"strings"
	"syscall/js"

	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/render"
	"github.com/nikolaydubina/jsonl-graph/web/svgpanzoom"
	"github.com/nikolaydubina/multiline-jsonl/mjsonl"
)

type layoutUpdater interface {
	UpdateGraphLayout(g render.Graph)
}

type Renderer struct {
	graphData     graph.Graph   // what graph contains
	graphRender   render.Graph  // how graph is rendered
	layoutUpdater layoutUpdater // how to make render graph
	containerID   string
	svgID         string
	rootID        string
	scaler        *svgpanzoom.PanZoomer
}

func NewRenderer(
	graphData graph.Graph,
	graphRender render.Graph,
	layoutUpdater layoutUpdater,
	containerID string,
	svgID string,
	rootID string,
	scaler *svgpanzoom.PanZoomer,
) Renderer {
	renderer := Renderer{
		graphData:     graphData,
		graphRender:   graphRender,
		layoutUpdater: layoutUpdater,
		containerID:   containerID,
		svgID:         svgID,
		rootID:        rootID,
		scaler:        scaler,
	}

	js.Global().Get("document").Call("getElementById", "inputData").Set("onkeyup", js.FuncOf(renderer.OnDataChange))
	js.Global().Get("document").Call("getElementById", "btnPrettifyJSON").Set("onclick", js.FuncOf(renderer.OnPrettifyJSON))
	js.Global().Get("document").Call("getElementById", "btnCollapseJSON").Set("onclick", js.FuncOf(renderer.OnCollapseJSON))
	js.Global().Get("document").Call("getElementById", "btnCollapseAllNodes").Set("onclick", js.FuncOf(renderer.OnCollapseAllNodes))
	js.Global().Get("document").Call("getElementById", "btnExpandAllNodes").Set("onclick", js.FuncOf(renderer.OnExpandAllNodes))

	return renderer
}

func (r Renderer) NewOnNodeTitleClickHandler(nodeTitleID string) func(_ js.Value, _ []js.Value) interface{} {
	return func(_ js.Value, _ []js.Value) interface{} {
		// natural id
		idParts := strings.Split(nodeTitleID, ":")
		id := strings.Join(idParts[4:], "")

		// internal id
		iid := r.graphData.IDStorage.Get(id)

		r.graphRender.Nodes[iid].ShowData = !r.graphRender.Nodes[iid].ShowData
		r.Render()
		return nil
	}
}

func (r Renderer) OnDataChange(_ js.Value, _ []js.Value) interface{} {
	inputString := js.Global().Get("document").Call("getElementById", "inputData").Get("value")

	g, err := graph.NewGraphFromJSONL(strings.NewReader(inputString.String()))
	if err != nil {
		log.Printf("bad input: %s", err)
		return nil
	}

	r.graphData.ReplaceFrom(g)
	r.Render()

	return nil
}

func (r Renderer) OnPrettifyJSON(_ js.Value, _ []js.Value) interface{} {
	inputString := js.Global().Get("document").Call("getElementById", "inputData").Get("value")

	var out bytes.Buffer
	if err := mjsonl.FormatJSONL(strings.NewReader(inputString.String()), &out, true); err != nil {
		log.Printf("bad input: %s", err)
		return nil
	}
	js.Global().Get("document").Call("getElementById", "inputData").Set("value", out.String())

	r.OnDataChange(js.Value{}, nil)
	return nil
}

func (r Renderer) OnCollapseJSON(_ js.Value, _ []js.Value) interface{} {
	inputString := js.Global().Get("document").Call("getElementById", "inputData").Get("value")

	var out bytes.Buffer
	if err := mjsonl.FormatJSONL(strings.NewReader(inputString.String()), &out, false); err != nil {
		log.Printf("bad input: %s", err)
		return nil
	}
	js.Global().Get("document").Call("getElementById", "inputData").Set("value", out.String())

	r.OnDataChange(js.Value{}, nil)
	return nil
}

func (r Renderer) OnCollapseAllNodes(_ js.Value, _ []js.Value) interface{} {
	for i := range r.graphRender.Nodes {
		r.graphRender.Nodes[i].ShowData = false
	}
	r.Render()
	return nil
}

func (r Renderer) OnExpandAllNodes(_ js.Value, _ []js.Value) interface{} {
	for i := range r.graphRender.Nodes {
		r.graphRender.Nodes[i].ShowData = true
	}
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
			r.graphRender.Nodes[id] = &render.Node{}
		}

		r.graphRender.Nodes[id].NodeData = node
		r.graphRender.Nodes[id].ID = node.ID()
		r.graphRender.Nodes[id].Title = node.ID()
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
			r.graphRender.Edges[fromID] = make(map[uint64]*render.Edge, len(edges))
		}

		// check all new data edges
		for toID := range edges {
			// new edge, creating new edge
			if _, ok := r.graphRender.Edges[fromID][toID]; !ok {
				r.graphRender.Edges[fromID][toID] = &render.Edge{}
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
		Call("getElementById", r.containerID).
		Set("innerHTML", r.graphRender.Render(r.svgID, r.rootID))

	for _, node := range r.graphRender.Nodes {
		nodeTitleID := node.NodeTitleID()
		js.Global().Get("document").Call("getElementById", nodeTitleID).Set("onclick", js.FuncOf(r.NewOnNodeTitleClickHandler(nodeTitleID)))
	}

	r.scaler.SetupHandlers()
	r.scaler.SetRootTranslation()
}

func main() {
	c := make(chan bool)

	renderer := NewRenderer(
		graph.NewGraph(),
		render.NewGraph(),
		render.BasicGridLayout{
			RowLength: 5,
			Margin:    25,
		},
		"output-container",
		"svg-jsonl-graph",
		"svg-jsonl-graph-root",
		svgpanzoom.NewPanZoomer(
			"svg-jsonl-graph",
			"svg-jsonl-graph-root",
			0.2,
		),
	)
	renderer.Render()
	renderer.OnDataChange(js.Value{}, nil)

	<-c
}
