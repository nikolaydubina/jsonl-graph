package render

import (
	"image"
)

// BasicGridLayout arranges nodes in rows and columns.
// Useful for debugging.
type BasicGridLayout struct {
	W         int
	H         int
	RowLength int
	Margin    int
}

func (l BasicGridLayout) UpdateGraphLayout(g Graph) {
	// update nodes positions
	i := 0
	for id, node := range g.Nodes {
		g.Nodes[id] = Node{
			Width:  l.W,
			Height: l.H,
			LeftBottom: image.Point{
				X: (i % l.RowLength) * (l.W + l.Margin),
				Y: (i / l.RowLength) * (l.W + l.Margin),
			},
			Title: node.Title,
		}
		i++
	}

	//  update edges
	for idFrom, toEdges := range g.Edges {
		for idTo := range toEdges {
			g.Edges[idFrom][idTo] = DirectEdge(g.Nodes[idFrom], g.Nodes[idTo])
		}
	}
}

func DirectEdge(from, to Node) Edge {
	return Edge{
		Points: []image.Point{
			from.LeftBottom.Add(image.Point{X: from.Width / 2, Y: from.Height / 2}),
			to.LeftBottom.Add(image.Point{X: to.Width / 2, Y: to.Height / 2}),
		},
	}
}
