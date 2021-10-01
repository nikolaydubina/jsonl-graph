package render

import (
	"log"
	"math"

	gnlayout "gonum.org/v1/gonum/graph/layout"
	gnsimple "gonum.org/v1/gonum/graph/simple"
	gnr2 "gonum.org/v1/gonum/spatial/r2"
)

func gonumNodeID(id uint64) int64 {
	return int64(float64(id))
}

func toGonumGraph(g Graph) *gnsimple.UndirectedGraph {
	gn := gnsimple.NewUndirectedGraph()
	for fromID, toIDs := range g.Edges {
		for toID := range toIDs {
			gn.SetEdge(gn.NewEdge(gnsimple.Node(gonumNodeID(fromID)), gnsimple.Node(gonumNodeID(toID))))
		}
	}
	return gn
}

type gnLayoutGetter interface {
	Coord2(id int64) gnr2.Vec
}

// will make dimension such that all nodes data fits into square of the same area
// this is for pretty layouts.
func getSquareLayoutSize(g Graph) float64 {
	s := g.TotalNodesWidth() * g.TotalNodesHeight()
	return math.Sqrt(float64(s))
}

func updateGraphByGonumLayout(g Graph, gnLayout gnLayoutGetter, scaleX float64, scaleY float64) {
	// get width and height of gonum layout
	gnw := 0.0
	gnh := 0.0
	for i := range g.Nodes {
		n := gnLayout.Coord2(gonumNodeID(i))

		if n.X > gnw {
			gnw = n.X
		}
		if n.Y > gnh {
			gnh = n.Y
		}
	}

	// get width and height of our expected layout
	w := getSquareLayoutSize(g) * scaleX
	h := w * scaleY
	log.Printf("gonum layout(%f x %f) our layout (%f x %f)", gnw, gnh, w, h)

	// update our coodinates and scale
	for nodeID := range g.Nodes {
		gnNode := gnLayout.Coord2(gonumNodeID(nodeID))

		x := gnNode.X * w / gnw
		y := gnNode.Y * h / gnh

		g.Nodes[nodeID].LeftBottom.X = int(x)
		g.Nodes[nodeID].LeftBottom.Y = int(y)
	}

	//  direct simple edges
	for idFrom, toEdges := range g.Edges {
		for idTo := range toEdges {
			edge := DirectEdge(*g.Nodes[idFrom], *g.Nodes[idTo])
			g.Edges[idFrom][idTo] = &edge
		}
	}

}

// This works, but not as pretty.
type EadesGonumLayout struct {
	Updates   int
	Repulsion float64
	Rate      float64
	Theta     float64
	ScaleX    float64
	ScaleY    float64
}

func (l EadesGonumLayout) UpdateGraphLayout(g Graph) {
	gn := toGonumGraph(g)

	eades := gnlayout.EadesR2{
		Updates:   l.Updates,
		Repulsion: l.Repulsion,
		Rate:      l.Rate,
		Theta:     l.Theta,
	}
	optimizer := gnlayout.NewOptimizerR2(gn, eades.Update)
	for optimizer.Update() {
	}

	updateGraphByGonumLayout(g, optimizer, l.ScaleX, l.ScaleY)
}

type IsomapR2GonumLayout struct {
	Scale  float64
	ScaleX float64
	ScaleY float64
}

func (l IsomapR2GonumLayout) UpdateGraphLayout(g Graph) {
	gn := toGonumGraph(g)
	optimizer := gnlayout.NewOptimizerR2(gn, gnlayout.IsomapR2{}.Update)
	for optimizer.Update() {
	}
	updateGraphByGonumLayout(g, optimizer, l.ScaleX, l.ScaleY)
}
