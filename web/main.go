package main

import (
	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/render"
	"github.com/nikolaydubina/jsonl-graph/web/renderer"
	"github.com/nikolaydubina/jsonl-graph/web/svgpanzoom"
)

func main() {
	c := make(chan bool)

	renderer.NewRenderer(
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

	<-c
}
