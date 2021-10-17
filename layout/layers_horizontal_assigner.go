package layout

import "github.com/nikolaydubina/jsonl-graph/layout/brandeskopf"

type BrandesKopfHorizontalAssigner struct {
	Delta int
}

func (s BrandesKopfHorizontalAssigner) AssignX(_ Graph, lg LayeredGraph) {
	blg := brandeskopf.LayeredGraph{
		Segments: lg.Segments,
		Dummy:    lg.Dummy,
		NodeYX:   lg.NodeYX,
		Layers:   lg.Layers(),
	}
	nodeX := brandeskopf.BrandesKopfLayersHorizontalAssignment(blg, s.Delta)
	for node, x := range nodeX {
		lg.NodeYX[node] = [2]int{lg.NodeYX[node][0], x}
	}
}
