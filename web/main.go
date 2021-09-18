package main

import (
	"log"
	"strings"
	"syscall/js"

	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/render"
)

// makeEmptyRenderGraph will make empty render graph based on graph.
func makeEmptyRenderGraph(g graph.Graph) render.Graph {
	gr := render.NewGraph()

	for id, node := range g.Nodes {
		gr.Nodes[id] = render.Node{
			Title: node.ID(),
		}
	}

	for fromID, toIDs := range g.Edges {
		gr.Edges[fromID] = make(map[uint64]render.Edge, len(toIDs))
		for toID, _ := range toIDs {
			gr.Edges[fromID][toID] = render.Edge{}
		}
	}

	return gr
}

// TODO: node location algorithm
// TODO: edge location algorithm
// TODO: set widths of nodes
// TODO: draw contents of nodes
// TODO: make small / make large nodes functions and callbacks
// TODO: coloring of nodes contents
// TODO: UI for coloring input

type Renderer struct{}

func (r *Renderer) Render() {
	input := `
{"id": "abcd"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"bufio"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"bytes"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"encoding/json"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"errors"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"fmt"}
{"from":"github.com/nikolaydubina/jsonl-graph/graph","to":"io"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"embed"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"encoding/json"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"errors"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"fmt"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"github.com/nikolaydubina/jsonl-graph/graph"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"image/color"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"io"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"io/ioutil"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"net/http"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"sort"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"strconv"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"strings"}
{"from":"github.com/nikolaydubina/jsonl-graph/dot","to":"text/template"}
{"from":"github.com/nikolaydubina/jsonl-graph","to":"flag"}
{"from":"github.com/nikolaydubina/jsonl-graph","to":"github.com/nikolaydubina/jsonl-graph/dot"}
{"from":"github.com/nikolaydubina/jsonl-graph","to":"github.com/nikolaydubina/jsonl-graph/graph"}
{"from":"github.com/nikolaydubina/jsonl-graph","to":"io"}
{"from":"github.com/nikolaydubina/jsonl-graph","to":"log"}
{"from":"github.com/nikolaydubina/jsonl-graph","to":"os"}
	`
	g, err := graph.NewGraphFromJSONLReader(strings.NewReader(input))
	if err != nil {
		log.Fatalf("expected no error but got %v", err)
	}

	gr := makeEmptyRenderGraph(g)

	layout := render.BasicGridLayout{
		W:         100,
		H:         16,
		RowLength: 10,
		Margin:    5,
	}
	layout.UpdateGraphLayout(gr)

	js.Global().
		Get("document").
		Call("getElementById", "output-container").
		Set("innerHTML", gr.Render())

	// TODO: avoid large JS code for zooming. use google/perf like zooming
	js.Global().
		Get("svgPanZoom").
		Invoke("#graph", map[string]interface{}{
			"minZoom":             0.1,
			"maxZoom":             10,
			"dblClickZoomEnabled": false,
		})
}

func main() {
	c := make(chan bool)

	renderer := Renderer{}
	renderer.Render()

	<-c
}
