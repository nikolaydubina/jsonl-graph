package renderer

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

// Renderer is a bridge between input, svg output, and browser controls.
// It registers handlers and invokes rendering.
// It combines all UI components together.
//
// Changes are of two types: structural — what is connected to what; and contents — node contents.
// Re-render on all changes.
// Re-layout on structural changes and big visual changes only.
type Renderer struct {
	graphData     graph.Graph   // what graph contains
	graphRender   render.Graph  // how graph is rendered
	layoutUpdater layoutUpdater // how to make render graph
	containerID   string
	svgID         string
	rootID        string
	scaler        *svgpanzoom.PanZoomer
	hasLayout     bool
	expandNodes   bool
}

func NewRenderer(
	graphData graph.Graph,
	graphRender render.Graph,
	layoutUpdater layoutUpdater,
	containerID string,
	svgID string,
	rootID string,
	scaler *svgpanzoom.PanZoomer,
) *Renderer {
	renderer := &Renderer{
		graphData:     graphData,
		graphRender:   graphRender,
		layoutUpdater: layoutUpdater,
		containerID:   containerID,
		svgID:         svgID,
		rootID:        rootID,
		scaler:        scaler,
		expandNodes:   false,
	}

	js.Global().Get("document").Call("getElementById", "inputData").Set("onkeyup", js.FuncOf(renderer.OnDataChange))
	js.Global().Get("document").Call("getElementById", "btnPrettifyJSON").Set("onclick", js.FuncOf(renderer.NewJSONFormatButtonHandler(true)))
	js.Global().Get("document").Call("getElementById", "btnCollapseJSON").Set("onclick", js.FuncOf(renderer.NewJSONFormatButtonHandler(false)))
	js.Global().Get("document").Call("getElementById", "switchExpandNodes").Set("onchange", js.FuncOf(renderer.SwitchExpandNodesHandler))

	layoutOptions := []LayoutOption{
		GridLayoutOption,
		ForcesLayoutOption,
		EadesLayoutOption,
		IsomapLayoutOption,
	}
	for _, l := range layoutOptions {
		js.Global().Get("document").Call("getElementById", string(l)).Set("onclick", js.FuncOf(renderer.NewLayoutOptionUpdater(l)))
	}

	renderer.OnDataChange(js.Value{}, nil)             // populating with data
	renderer.SwitchExpandNodesHandler(js.Value{}, nil) // expanding nodes

	return renderer
}

func (r *Renderer) NewOnNodeTitleClickHandler(nodeTitleID string) func(_ js.Value, _ []js.Value) interface{} {
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

func (r *Renderer) OnDataChange(_ js.Value, _ []js.Value) interface{} {
	tracker := graph.NewGraphTracker(r.graphData)

	inputString := js.Global().Get("document").Call("getElementById", "inputData").Get("value")
	g, err := graph.NewGraphFromJSONL(strings.NewReader(inputString.String()))
	if err != nil {
		log.Printf("bad input: %s", err)
		return nil
	}

	r.graphData.ReplaceFrom(g)
	UpdateRenderGraphWithDataGraph(r.graphData, r.graphRender)

	// update layout only on structural changes.
	if tracker.HasChanged(r.graphData) {
		r.layoutUpdater.UpdateGraphLayout(r.graphRender)
		CenterGraph(r.graphRender, r.scaler)
	}

	r.Render()
	return nil
}

func (r *Renderer) NewJSONFormatButtonHandler(prettify bool) func(_ js.Value, _ []js.Value) interface{} {
	return func(_ js.Value, _ []js.Value) interface{} {
		inputString := js.Global().Get("document").Call("getElementById", "inputData").Get("value")

		var out bytes.Buffer
		if err := mjsonl.FormatJSONL(strings.NewReader(inputString.String()), &out, prettify); err != nil {
			log.Printf("bad input: %s", err)
			return nil
		}
		js.Global().Get("document").Call("getElementById", "inputData").Set("value", out.String())

		r.OnDataChange(js.Value{}, nil)
		return nil
	}
}

// collapsing or expanding all nodes changes graph a lot, so re-copmuting layout
func (r *Renderer) SwitchExpandNodesHandler(_ js.Value, _ []js.Value) interface{} {
	r.expandNodes = !r.expandNodes

	for i := range r.graphRender.Nodes {
		r.graphRender.Nodes[i].ShowData = r.expandNodes
	}
	r.layoutUpdater.UpdateGraphLayout(r.graphRender)
	CenterGraph(r.graphRender, r.scaler)
	r.Render()
	return nil
}

func (r *Renderer) Render() {
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
