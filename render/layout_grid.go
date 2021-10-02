package render

import (
	"image"
	"math/rand"
	"sort"
)

// randomize positions
func InitRandom(g Graph) {
	d := getSquareLayoutSize(g)
	for i := range g.Nodes {
		g.Nodes[i].LeftBottom.X = int(rand.Float64() * d)
		g.Nodes[i].LeftBottom.Y = int(rand.Float64() * d)
	}
}

// BasicGridLayout arranges nodes in rows and columns.
// Useful for debugging.
type BasicGridLayout struct {
	RowLength int
	Margin    int
}

func (l BasicGridLayout) UpdateGraphLayout(g Graph) {
	// sort nodes by id
	nodes := make([]uint64, 0, len(g.Nodes))
	for id := range g.Nodes {
		nodes = append(nodes, id)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i] < nodes[j] })

	// update nodes positions
	x := 0
	y := 0
	yf := 0
	for i, id := range nodes {
		g.Nodes[id].LeftBottom = image.Point{X: x, Y: y}

		colNum := (i + 1) % l.RowLength

		if colNum == 0 {
			x = 0
			y += yf + l.Margin
			yf = g.Nodes[id].Height()
		} else {
			x += g.Nodes[id].Width() + l.Margin
			if h := g.Nodes[id].Height(); h > yf {
				yf = h
			}
		}
	}

	//  update edges
	for idFrom, toEdges := range g.Edges {
		for idTo := range toEdges {
			edge := DirectEdge(*g.Nodes[idFrom], *g.Nodes[idTo])
			g.Edges[idFrom][idTo] = &edge
		}
	}
}
