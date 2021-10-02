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

// Renderer bridge between input, svg output, and browser controls.
// Changes are of two types: structural (what is connected to what) and contents (node contents).
// We re-compute layout when structural changes or large enough visual changes.
// We do not re-compute layout on content changes.
type Renderer struct {
	graphData     graph.Graph   // what graph contains
	graphRender   render.Graph  // how graph is rendered
	layoutUpdater layoutUpdater // how to make render graph
	containerID   string
	svgID         string
	rootID        string
	scaler        *svgpanzoom.PanZoomer
	hasLayout     bool
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
	}

	js.Global().Get("document").Call("getElementById", "inputData").Set("onkeyup", js.FuncOf(renderer.OnDataChange))
	js.Global().Get("document").Call("getElementById", "btnPrettifyJSON").Set("onclick", js.FuncOf(renderer.NewJSONFormatButtonHandler(true)))
	js.Global().Get("document").Call("getElementById", "btnCollapseJSON").Set("onclick", js.FuncOf(renderer.NewJSONFormatButtonHandler(false)))
	js.Global().Get("document").Call("getElementById", "btnCollapseAllNodes").Set("onclick", js.FuncOf(renderer.OnCollapseAllNodes))
	js.Global().Get("document").Call("getElementById", "btnExpandAllNodes").Set("onclick", js.FuncOf(renderer.OnExpandAllNodes))

	layoutOptions := []LayoutOption{
		GridLayoutOption,
		ForcesLayoutOption,
		EadesLayoutOption,
		IsomapLayoutOption,
	}
	for _, l := range layoutOptions {
		js.Global().Get("document").Call("getElementById", string(l)).Set("onclick", js.FuncOf(renderer.NewLayoutOptionUpdater(l)))
	}

	return renderer
}

type LayoutOption string

const (
	GridLayoutOption   LayoutOption = "layoutOptionGrid"
	ForcesLayoutOption LayoutOption = "layoutOptionForces"
	EadesLayoutOption  LayoutOption = "layoutOptionEades"
	IsomapLayoutOption LayoutOption = "layoutOptionIsomap"
)

// TODO: read options of layout from UI
func (r *Renderer) NewLayoutOptionUpdater(layoutOption LayoutOption) func(_ js.Value, _ []js.Value) interface{} {
	return func(_ js.Value, _ []js.Value) interface{} {
		switch layoutOption {
		case GridLayoutOption:
			r.layoutUpdater = render.BasicGridLayout{
				RowLength: 5,
				Margin:    25,
			}
		case ForcesLayoutOption:
			render.InitRandom(r.graphRender)
			r.layoutUpdater = render.ForceGraphLayout{
				Delta:    1,
				MaxSteps: 5000,
				Epsilon:  1.5,
				Forces: []render.Force{
					render.GravityForce{
						K:         -50,
						EdgesOnly: false,
					},
					render.SpringForce{
						K:         0.2,
						L:         200,
						EdgesOnly: true,
					},
				},
			}
		case EadesLayoutOption:
			r.layoutUpdater = render.EadesGonumLayout{
				Repulsion: 1,
				Rate:      0.05,
				Updates:   30,
				Theta:     0.2,
				ScaleX:    0.5,
				ScaleY:    0.5,
			}
		case IsomapLayoutOption:
			r.layoutUpdater = render.IsomapR2GonumLayout{
				ScaleX: 0.5,
				ScaleY: 0.5,
			}
		}

		r.layoutUpdater.UpdateGraphLayout(r.graphRender)
		centerGraph(r.graphRender, r.scaler)
		r.Render()
		return nil
	}
}

// centerGraph will reset transformations, center it and apply zoom.
func centerGraph(g render.Graph, scaler *svgpanzoom.PanZoomer) {
	wScreen := js.Global().Get("document").Call("width")
	hScreen := js.Global().Get("document").Call("height")

	wGraph := float64(g.Width())
	hGraph := float64(g.Height())

	log.Printf("screen (%f x %f) graph (%f x %f)", wScreen, hScreen, wGraph, hGraph)

	dx := 0.0
	dy := 0.0

	scaler.Reset().Shift(dx, dy).Zoom(1)
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
		centerGraph(r.graphRender, r.scaler)
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

// collapsing all nodes changes graph a lot, so re-copmuting layout
func (r *Renderer) OnCollapseAllNodes(_ js.Value, _ []js.Value) interface{} {
	for i := range r.graphRender.Nodes {
		r.graphRender.Nodes[i].ShowData = false
	}
	r.layoutUpdater.UpdateGraphLayout(r.graphRender)
	centerGraph(r.graphRender, r.scaler)
	r.Render()
	return nil
}

// expanding all nodes changes graph a lot, so re-copmuting layout
func (r *Renderer) OnExpandAllNodes(_ js.Value, _ []js.Value) interface{} {
	for i := range r.graphRender.Nodes {
		r.graphRender.Nodes[i].ShowData = true
	}
	r.layoutUpdater.UpdateGraphLayout(r.graphRender)
	centerGraph(r.graphRender, r.scaler)
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
