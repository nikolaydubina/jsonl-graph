package app

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"syscall/js"

	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/layout"
	"github.com/nikolaydubina/jsonl-graph/render"
	"github.com/nikolaydubina/jsonl-graph/web/svgpanzoom"
	"github.com/nikolaydubina/multiline-jsonl/mjsonl"
)

// Bridge between input, svg output, and browser controls.
type Bridge struct {
	graphData     graph.Graph           // what graph contains
	graphLayout   layout.Graph          // how nodes located and what are edge paths
	layoutUpdater layout.Layout         // how to arrange graph
	expandNodes   map[uint64]bool       // which nodes to expand
	scaler        *svgpanzoom.PanZoomer // how to scale and zoom svg
	scalerLayout  layout.MemoLayout     // how distance between nodes is done for given layout
	containerID   string
	svgID         string
	rootID        string
	prettifyJSON  bool
}

func NewBridge(
	containerID string,
	svgID string,
	rootID string,
) *Bridge {
	graphLayout := layout.Graph{}

	renderer := Bridge{
		graphData:     graph.NewGraph(),
		graphLayout:   graphLayout,
		layoutUpdater: layout.NewBasicSugiyamaLayersGraphLayout(),
		containerID:   containerID,
		svgID:         svgID,
		rootID:        rootID,
		scaler: svgpanzoom.NewPanZoomer(
			svgID,
			rootID,
			0.2,
		),
		scalerLayout: layout.MemoLayout{
			Layout: &layout.ScalerLayout{Scale: 1},
			Graph:  graphLayout,
		},
	}

	document := js.Global().Get("document")

	document.Call("getElementById", "inputData").Set("onkeyup", js.FuncOf(renderer.OnDataChangeHandler))
	document.Call("getElementById", "switchPrettifyJSON").Set("onchange", js.FuncOf(renderer.SwitchPrettifyJSONHandler))
	document.Call("getElementById", "switchExpandNodes").Set("onchange", js.FuncOf(renderer.SwitchExpandNodesHandler))
	document.Call("getElementById", "rangeNodeDistance").Set("oninput", js.FuncOf(renderer.NodeDistanceRangeHandler))

	for _, l := range AllLayoutOptions() {
		document.Call("getElementById", string(l)).Set("onclick", js.FuncOf(renderer.NewLayoutOptionUpdater(l)))
	}

	renderer.OnDataChangeHandler(js.Value{}, nil)      // populating with data
	renderer.SwitchExpandNodesHandler(js.Value{}, nil) // expanding nodes

	return &renderer
}

// newExpandAllNodesForGraph will make expand node tracking structure with all nodes expanded for graph.
func newExpandAllNodesForGraph(g graph.Graph) map[uint64]bool {
	nodes := make(map[uint64]bool, len(g.Nodes))
	for n := range g.Nodes {
		nodes[n] = true
	}
	return nodes
}

func (r *Bridge) OnDataChangeHandler(_ js.Value, _ []js.Value) interface{} {
	tracker := graph.NewGraphTracker(r.graphData)

	input := js.Global().Get("document").Call("getElementById", "inputData").Get("value").String()
	g, err := graph.NewGraphFromJSONL(strings.NewReader(input))
	if err != nil {
		log.Printf("can not get new graph data: %s", err)
		return nil
	}

	r.graphData.ReplaceFrom(g)
	log.Printf("got new graph data: %s\n", r.graphData)

	if tracker.HasStructureChanged(r.graphData) {
		// compute w and h for nodes
		rgraph := render.NewGraph()
		for id, node := range r.graphData.Nodes {
			rgraph.Nodes[id] = render.Node{
				NodeData: node,
			}
		}

		// update graph layout graph
		r.graphLayout = layout.Graph{
			Nodes: make(map[uint64]layout.Node, len(r.graphData.Nodes)),
			Edges: make(map[[2]uint64]layout.Edge, len(r.graphData.Edges)),
		}
		for id := range r.graphData.Nodes {
			r.graphLayout.Nodes[id] = layout.Node{
				W: rgraph.Nodes[id].Width(),
				H: rgraph.Nodes[id].Height(),
			}
		}
		for e := range r.graphData.Edges {
			r.graphLayout.Edges[e] = layout.Edge{}
		}

		// expand nodes
		if r.expandNodes == nil {
			r.expandNodes = newExpandAllNodesForGraph(r.graphData)
		}

		r.layoutUpdater.UpdateGraphLayout(r.graphLayout)
		r.scalerLayout.Graph = layout.CopyGraph(r.graphLayout) // memoize for scaling
		r.CenterGraph()
	}

	r.Render()
	return nil
}

func (r *Bridge) NewOnNodeTitleClickHandler(id uint64) func(_ js.Value, _ []js.Value) interface{} {
	return func(_ js.Value, _ []js.Value) interface{} {
		r.expandNodes[id] = !r.expandNodes[id]
		r.Render()
		return nil
	}
}

func (r *Bridge) NodeDistanceRangeHandler(_ js.Value, args []js.Value) interface{} {
	rawval := args[0].Get("target").Get("value").String()
	val := 1.0
	if n, err := fmt.Sscanf(rawval, "%f", &val); n != 1 || err != nil {
		log.Printf("handler: node distance range: error: %s", err)
		return nil
	}

	// updating parameter for scaling
	if v, ok := r.scalerLayout.Layout.(*layout.ScalerLayout); ok {
		v.Scale = val
	}

	// only running memoized scaling layout
	r.scalerLayout.UpdateGraphLayout(r.graphLayout)

	r.Render()
	return nil
}

func (r *Bridge) SwitchPrettifyJSONHandler(_ js.Value, _ []js.Value) interface{} {
	r.prettifyJSON = !r.prettifyJSON

	inputString := js.Global().Get("document").Call("getElementById", "inputData").Get("value").String()

	var out bytes.Buffer
	if err := mjsonl.FormatJSONL(strings.NewReader(inputString), &out, r.prettifyJSON); err != nil {
		log.Printf("bad input: %s", err)
		return nil
	}
	js.Global().Get("document").Call("getElementById", "inputData").Set("value", out.String())

	r.OnDataChangeHandler(js.Value{}, nil)
	return nil
}

// collapsing or expanding all nodes changes graph a lot, so re-copmuting layout
func (r *Bridge) SwitchExpandNodesHandler(_ js.Value, _ []js.Value) interface{} {
	r.expandNodes = newExpandAllNodesForGraph(r.graphData)
	r.SetInitialUpdateGraphLayout()
	r.Render()
	return nil
}

func (r *Bridge) CenterGraph() {
	minx, miny, maxx, maxy := layout.BoundingBox(r.graphLayout)
	r.scaler.CenterBox(float64(minx), float64(miny), float64(maxx), float64(maxy))
}

func (r *Bridge) Render() {
	graph := render.NewGraph()
	graph.ID = r.rootID

	// add nodes data
	// add nodes positions
	for id, node := range r.graphData.Nodes {
		graph.Nodes[id] = render.Node{
			ID:       fmt.Sprintf("%d", id),
			XY:       r.graphLayout.Nodes[id].XY,
			Title:    node.ID(),
			NodeData: node,
		}
	}

	// update graph layout graph
	for e, edata := range r.graphLayout.Edges {
		graph.Edges[e] = render.Edge{
			Path: edata.Path,
		}
	}

	svgContainer := render.SVG{
		ID: r.svgID,
		Definitions: []render.Renderable{
			render.ArrowDef{},
		},
		Body: graph,
	}

	js.Global().Get("document").Call("getElementById", r.containerID).Set("innerHTML", svgContainer.Render())

	for id, node := range graph.Nodes {
		js.Global().Get("document").Call("getElementById", node.TitleID()).Set("onclick", js.FuncOf(r.NewOnNodeTitleClickHandler(id)))
	}

	r.scaler.SetupHandlers()
	r.scaler.SetRootTranslation()
}
