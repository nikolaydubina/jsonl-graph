package render

import (
	"image"
	"sort"
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
	// sort nodes by id
	nodes := make([]uint64, 0, len(g.Nodes))
	for id := range g.Nodes {
		nodes = append(nodes, id)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i] < nodes[j] })

	// update nodes positions
	i := 0
	for _, id := range nodes {
		g.Nodes[id] = Node{
			Width:  l.W,
			Height: l.H,
			LeftBottom: image.Point{
				X: (i % l.RowLength) * (l.W + l.Margin),
				Y: (i / l.RowLength) * (l.W + l.Margin),
			},
			Title: g.Nodes[id].Title,
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
