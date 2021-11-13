package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"syscall/js"

	"github.com/nikolaydubina/go-svgpan"
	"github.com/nikolaydubina/multiline-jsonl/mjsonl"

	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/layout"
	"github.com/nikolaydubina/jsonl-graph/layout/brandeskopf"
	"github.com/nikolaydubina/jsonl-graph/svg"
)

// Bridge between browser input, layout computation, and browser svg output.
type Bridge struct {
	graphData        graph.Graph          // what graph contains
	graphLayout      layout.Graph         // how nodes located and what are edge paths
	layoutUpdater    layout.Layout        // how to arrange graph
	expandNodeSwitch bool                 // value of expand all nodes switch
	prettifyJSON     bool                 // format JSON input
	expandNodes      map[uint64]bool      // which nodes to expand
	scaler           *svgpan.SVGPanZoomer // how to scale and zoom svg
	scalerLayout     MemoLayout           // how distance between nodes is done for given layout
	graphRenderer    GraphRenderer        // how to set graph SVG into HTML and bind handlers
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

	// update nodes and add new ones
	for id, node := range r.graphData.Nodes {
		// compute w and h for nodes, since width and height of node depends on content
		rnodeData := node
		if !r.expandNodes[id] {
			rnodeData = nil
		}
		rnode := svg.Node{
			Title:    node.ID(),
			NodeData: rnodeData,
		}
		w := rnode.Width()
		h := rnode.Height()

		r.graphLayout.Nodes[id] = layout.Node{W: w, H: h}
	}

	// remove old nodes
	for id := range r.graphLayout.Nodes {
		if _, ok := r.graphData.Nodes[id]; !ok {
			delete(r.graphLayout.Nodes, id)
		}
	}

	// add new edges
	for e := range r.graphData.Edges {
		r.graphLayout.Edges[e] = layout.Edge{}
	}

	// remove non existent edges
	for e := range r.graphLayout.Edges {
		if _, ok := r.graphData.Edges[e]; !ok {
			delete(r.graphLayout.Edges, e)
		}
	}

	if tracker.HasStructureChanged(r.graphData) {
		// expand nodes
		if r.expandNodes == nil {
			r.expandNodes = newExpandAllNodesForGraph(r.graphData)
		}

		r.layoutUpdater.UpdateGraphLayout(r.graphLayout)
		r.scalerLayout.Graph = r.graphLayout.Copy() // memoize for scaling
		r.CenterGraph()
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
	return nil
}

func (r *Bridge) NewOnNodeTitleClickHandler(id uint64) func(_ js.Value, _ []js.Value) interface{} {
	return func(_ js.Value, _ []js.Value) interface{} {
		r.expandNodes[id] = !r.expandNodes[id]
		r.Render()
		return nil
	}
}

// collapsing or expanding all nodes changes graph a lot, so re-copmuting layout
func (r *Bridge) SwitchExpandNodesHandler(_ js.Value, e []js.Value) interface{} {
	r.expandNodeSwitch = !r.expandNodeSwitch
	for k := range r.expandNodes {
		r.expandNodes[k] = r.expandNodeSwitch
	}
	r.SetInitialUpdateGraphLayout()
	r.Render()
	return nil
}

func (r *Bridge) CenterGraph() {
	minx, miny, maxx, maxy := r.graphLayout.BoundingBox()
	r.scaler.CenterBox(float64(minx), float64(miny), float64(maxx), float64(maxy))
}

func (r *Bridge) Render() {
	r.graphRenderer.RenderGraph(r.graphData, r.graphLayout, r.expandNodes)
	r.scaler.SetupHandlers()
	r.scaler.SetRootTranslation()
}

// GraphRenderer renders graph data based on layout and display options into HTML and binds JS handlers.
type GraphRenderer struct {
	containerID                      string
	svgID                            string
	rootID                           string
	nodeTitleClickHandlerConstructor func(id uint64) func(_ js.Value, _ []js.Value) interface{}
}

func (r GraphRenderer) RenderGraph(g graph.Graph, gl layout.Graph, expandNodes map[uint64]bool) {
	graph := svg.NewGraph()
	graph.ID = r.rootID

	// add nodes data and positions
	for id, node := range g.Nodes {
		nodeData := node
		if !expandNodes[id] {
			nodeData = nil
		}
		graph.Nodes[id] = svg.Node{
			ID:       fmt.Sprintf("%d", id),
			XY:       gl.Nodes[id].XY,
			Title:    node.ID(),
			NodeData: nodeData,
		}
	}

	// update graph layout graph
	for e, edata := range gl.Edges {
		graph.Edges[e] = svg.Edge{
			Path: edata.Path,
		}
	}

	svgContainer := svg.SVG{
		ID: r.svgID,
		Definitions: []svg.Renderable{
			svg.ArrowDef{},
		},
		Body: graph,
	}

	// set content of HTML
	js.Global().Get("document").Call("getElementById", r.containerID).Set("innerHTML", svgContainer.Render())

	// bind handlers
	for id, node := range graph.Nodes {
		js.Global().Get("document").Call("getElementById", node.TitleID()).Set("onclick", js.FuncOf(r.nodeTitleClickHandlerConstructor(id)))
	}
}

type LayoutOption string

const (
	ForcesLayoutOption LayoutOption = "layoutOptionForces"
	EadesLayoutOption  LayoutOption = "layoutOptionEades"
	IsomapLayoutOption LayoutOption = "layoutOptionIsomap"
	LayersLayoutOption LayoutOption = "layoutOptionLayers"
)

func AllLayoutOptions() []LayoutOption {
	return []LayoutOption{
		ForcesLayoutOption,
		EadesLayoutOption,
		IsomapLayoutOption,
		LayersLayoutOption,
	}
}

func (r *Bridge) NewLayoutOptionUpdater(layoutOption LayoutOption) func(_ js.Value, _ []js.Value) interface{} {
	return func(_ js.Value, _ []js.Value) interface{} {
		switch layoutOption {
		case ForcesLayoutOption:
			r.layoutUpdater = layout.SequenceLayout{
				Layouts: []layout.Layout{
					layout.ForceGraphLayout{
						Delta:    1,
						MaxSteps: 5000,
						Epsilon:  1.5,
						Forces: []layout.Force{
							layout.GravityForce{
								K:         -50,
								EdgesOnly: false,
							},
							layout.SpringForce{
								K:         0.2,
								L:         200,
								EdgesOnly: true,
							},
						},
					},
					layout.DirectEdgesLayout{},
				},
			}
		case EadesLayoutOption:
			r.layoutUpdater = layout.SequenceLayout{
				Layouts: []layout.Layout{
					layout.EadesGonumLayout{
						Repulsion: 1,
						Rate:      0.05,
						Updates:   30,
						Theta:     0.2,
						ScaleX:    0.5,
						ScaleY:    0.5,
					},
					layout.DirectEdgesLayout{},
				},
			}
		case IsomapLayoutOption:
			r.layoutUpdater = layout.SequenceLayout{
				Layouts: []layout.Layout{
					layout.IsomapR2GonumLayout{
						ScaleX: 0.5,
						ScaleY: 0.5,
					},
					layout.DirectEdgesLayout{},
				},
			}
		case LayersLayoutOption:
			r.layoutUpdater = layout.SugiyamaLayersStrategyGraphLayout{
				CycleRemover:   layout.NewSimpleCycleRemover(),
				LevelsAssigner: layout.NewLayeredGraph,
				OrderingAssigner: layout.LBLOrderingOptimizer{
					Epochs: 10,
					LayerOrderingOptimizer: layout.RandomLayerOrderingOptimizer{
						Epochs: 5,
					},
				}.Optimize,
				NodesHorizontalCoordinatesAssigner: brandeskopf.BrandesKopfLayersNodesHorizontalAssigner{
					Delta: 25, // TODO: get dynamically from graph width
				},
				EdgePathAssigner: layout.StraightEdgePathAssigner{
					MarginY:        25,
					MarginX:        25,
					FakeNodeWidth:  25,
					FakeNodeHeight: 25,
				}.UpdateGraphLayout,
			}
		default:
			log.Printf("unexpected layout option(%s)", layoutOption)
		}

		r.SetInitialUpdateGraphLayout()
		r.Render()
		return nil
	}
}

// SetInitialUpdateGraphLayout sets layout to what it should look like at the begging for a layout.
// TODO: add edge direct path layout to basic layouts
func (r *Bridge) SetInitialUpdateGraphLayout() {
	r.layoutUpdater.UpdateGraphLayout(r.graphLayout)
	r.scalerLayout.Graph = r.graphLayout.Copy()
	r.scalerLayout.UpdateGraphLayout(r.graphLayout)
	r.CenterGraph()
}

// MemoLayout applies layout updaters to memoized graph and stores to destination.
type MemoLayout struct {
	Graph  layout.Graph
	Layout layout.Layout
}

func (l MemoLayout) UpdateGraphLayout(g layout.Graph) {
	newgraph := l.Graph.Copy()
	l.Layout.UpdateGraphLayout(newgraph)

	// apply to target graph
	for i := range g.Nodes {
		g.Nodes[i] = layout.Node{
			XY: newgraph.Nodes[i].XY,
			W:  g.Nodes[i].W,
			H:  g.Nodes[i].H,
		}
	}
	for e := range g.Edges {
		g.Edges[e] = layout.Edge{Path: make([][2]int, len(newgraph.Edges[e].Path))}
		for i, ne := range newgraph.Edges[e].Path {
			g.Edges[e].Path[i] = ne
		}
	}
}

func main() {
	containerID := "output-container"
	svgID := "svg-jsonl-graph"
	rootID := "svg-jsonl-graph-root"

	graphLayout := layout.Graph{
		Nodes: make(map[uint64]layout.Node),
		Edges: make(map[[2]uint64]layout.Edge),
	}

	renderer := Bridge{
		graphData:   graph.NewGraph(),
		graphLayout: graphLayout,
		scaler: svgpan.NewSVGPanZoomer(
			svgID,
			rootID,
			0.2,
		),
		scalerLayout: MemoLayout{
			Layout: &layout.ScalerLayout{
				Scale:      1,
				EdgeLayout: layout.DirectEdgesLayout{},
			},
			Graph: graphLayout,
		},
		expandNodeSwitch: false, // deafult is true, switching to true bellow after data is loaded
	}
	graphRenderer := GraphRenderer{
		containerID:                      containerID,
		svgID:                            svgID,
		rootID:                           rootID,
		nodeTitleClickHandlerConstructor: renderer.NewOnNodeTitleClickHandler,
	}
	renderer.graphRenderer = graphRenderer

	document := js.Global().Get("document")

	document.Call("getElementById", "inputData").Set("onkeyup", js.FuncOf(renderer.OnDataChangeHandler))
	document.Call("getElementById", "switchPrettifyJSON").Set("onchange", js.FuncOf(renderer.SwitchPrettifyJSONHandler))
	document.Call("getElementById", "switchExpandNodes").Set("onchange", js.FuncOf(renderer.SwitchExpandNodesHandler))
	document.Call("getElementById", "rangeNodeDistance").Set("oninput", js.FuncOf(renderer.NodeDistanceRangeHandler))

	for _, l := range AllLayoutOptions() {
		document.Call("getElementById", string(l)).Set("onclick", js.FuncOf(renderer.NewLayoutOptionUpdater(l)))
	}

	// invoke events on start
	renderer.NewLayoutOptionUpdater(LayersLayoutOption)(js.Value{}, nil) // setting default updater
	renderer.OnDataChangeHandler(js.Value{}, nil)                        // populating with data
	renderer.SwitchExpandNodesHandler(js.Value{}, nil)                   // expanding nodes

	// do not exit
	c := make(chan bool)
	<-c
}
