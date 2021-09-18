package render

import "image"

// BasicGridLayout arranges nodes in rows and columns.
// Useful for debugging.
type BasicGridLayout struct {
	W         int
	H         int
	RowLength int
	Margin    int
}

func (l BasicGridLayout) UpdateGraphLayout(g Graph) {
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
}
