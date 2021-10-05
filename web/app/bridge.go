package app

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"syscall/js"

	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/render"
	"github.com/nikolaydubina/jsonl-graph/web/svgpanzoom"
	"github.com/nikolaydubina/multiline-jsonl/mjsonl"
)

// Bridge between input, svg output, and browser controls.
// It registers handlers and invokes rendering.
// It combines all UI components together.
//
// Changes are of two types: structural — what is connected to what; and contents — node contents.
// Re-render on all changes.
// Re-layout on structural changes and big visual changes only.
type Bridge struct {
	graphData     graph.Graph   // what graph contains
	graphRender   render.Graph  // how graph is rendered
	layoutUpdater render.Layout // how to arrange graph
	containerID   string
	svgID         string
	rootID        string
	scaler        *svgpanzoom.PanZoomer
	hasLayout     bool
	expandNodes   bool
	prettifyJSON  bool
	scalerLayout  render.MemoLayout
}

func NewBridge(
	graphData graph.Graph,
	graphRender render.Graph,
	containerID string,
	svgID string,
	rootID string,
	scaler *svgpanzoom.PanZoomer,
) *Bridge {
	scalerLayout := render.ScalerLayout{Scale: 1}

	renderer := &Bridge{
		graphData:   graphData,
		graphRender: graphRender,
		layoutUpdater: render.CompositeLayout{
			Layouts: []render.Layout{
				render.BasicLayersLayout{
					MarginX:        25,
					MarginY:        25,
					FakeNodeWidth:  25,
					FakeNodeHeight: 25,
				},
				&scalerLayout,
			},
		},
		containerID: containerID,
		svgID:       svgID,
		rootID:      rootID,
		scaler:      scaler,
		expandNodes: false,
		scalerLayout: render.MemoLayout{
			Layout: &scalerLayout,
			Graph:  graphRender,
		},
	}

	document := js.Global().Get("document")

	document.Call("getElementById", "inputData").Set("onkeyup", js.FuncOf(renderer.OnDataChange))
	document.Call("getElementById", "switchPrettifyJSON").Set("onchange", js.FuncOf(renderer.SwitchPrettifyJSONHandler))
	document.Call("getElementById", "switchExpandNodes").Set("onchange", js.FuncOf(renderer.SwitchExpandNodesHandler))
	document.Call("getElementById", "rangeNodeDistance").Set("oninput", js.FuncOf(renderer.NodeDistanceRangeHandler))

	for _, l := range AllLayoutOptions() {
		document.Call("getElementById", string(l)).Set("onclick", js.FuncOf(renderer.NewLayoutOptionUpdater(l)))
	}

	renderer.OnDataChange(js.Value{}, nil)             // populating with data
	renderer.SwitchExpandNodesHandler(js.Value{}, nil) // expanding nodes

	return renderer
}

func (r *Bridge) NewOnNodeTitleClickHandler(nodeTitleID string) func(_ js.Value, _ []js.Value) interface{} {
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

func (r *Bridge) OnDataChange(_ js.Value, _ []js.Value) interface{} {
	tracker := graph.NewGraphTracker(r.graphData)

	inputString := js.Global().Get("document").Call("getElementById", "inputData").Get("value")
	g, err := graph.NewGraphFromJSONL(strings.NewReader(inputString.String()))
	if err != nil {
		log.Printf("bad input: %s", err)
		return nil
	}
	log.Printf("new graph data: %s", g)

	r.graphData.ReplaceFrom(g)
	log.Printf("new graph data: %s", r.graphData)
	UpdateRenderGraphWithDataGraph(r.graphData, r.graphRender)

	// update layout only on structural changes.
	if tracker.HasChanged(r.graphData) {
		r.layoutUpdater.UpdateGraphLayout(r.graphRender)
		log.Printf("new graph layout: %s", r.graphRender)
		r.scalerLayout.Graph = r.graphRender.Copy() // memoize for scaling
		CenterGraph(r.graphRender, r.scaler)
	}

	r.Render()
	return nil
}

func (r *Bridge) NodeDistanceRangeHandler(_ js.Value, args []js.Value) interface{} {
	rawval := args[0].Get("target").Get("value").String()
	val := 1.0
	if n, err := fmt.Sscanf(rawval, "%f", &val); n != 1 || err != nil {
		log.Printf("handler: node distance range: error: %s", err)
		return nil
	}

	log.Printf("handler: node distance range: new value(%f)", val)

	// updating parameter for scaling
	if v, ok := r.scalerLayout.Layout.(*render.ScalerLayout); ok {
		v.Scale = val
	}

	// only running memoized scaling layout
	r.scalerLayout.UpdateGraphLayout(r.graphRender)

	r.Render()
	return nil
}

func (r *Bridge) SwitchPrettifyJSONHandler(_ js.Value, _ []js.Value) interface{} {
	r.prettifyJSON = !r.prettifyJSON

	inputString := js.Global().Get("document").Call("getElementById", "inputData").Get("value")

	var out bytes.Buffer
	if err := mjsonl.FormatJSONL(strings.NewReader(inputString.String()), &out, r.prettifyJSON); err != nil {
		log.Printf("bad input: %s", err)
		return nil
	}
	js.Global().Get("document").Call("getElementById", "inputData").Set("value", out.String())

	r.OnDataChange(js.Value{}, nil)
	return nil
}

// collapsing or expanding all nodes changes graph a lot, so re-copmuting layout
func (r *Bridge) SwitchExpandNodesHandler(_ js.Value, _ []js.Value) interface{} {
	r.expandNodes = !r.expandNodes

	for i := range r.graphRender.Nodes {
		r.graphRender.Nodes[i].ShowData = r.expandNodes
	}
	r.layoutUpdater.UpdateGraphLayout(r.graphRender)
	r.scalerLayout.Graph = r.graphRender.Copy() // memoize for scaling
	CenterGraph(r.graphRender, r.scaler)
	r.Render()
	return nil
}

func (r *Bridge) Render() {
	document := js.Global().Get("document")
	document.Call("getElementById", r.containerID).Set("innerHTML", r.graphRender.Render(r.svgID, r.rootID))

	for _, node := range r.graphRender.Nodes {
		document.Call("getElementById", node.TitleID()).Set("onclick", js.FuncOf(r.NewOnNodeTitleClickHandler(node.TitleID())))
	}

	r.scaler.SetupHandlers()
	r.scaler.SetRootTranslation()
}
