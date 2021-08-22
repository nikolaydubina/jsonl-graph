package render

import "image"

// MakeBasicGridLayout will set some random layout for development
func MakeBasicGridLayout(g Graph) {
	w := 100
	h := 16

	i := 0
	rowLength := 10
	margin := 5

	for id, node := range g.Nodes {
		g.Nodes[id] = Node{
			Width:  w,
			Height: h,
			LeftBottom: image.Point{
				X: (i % rowLength) * (w + margin),
				Y: (i / rowLength) * (w + margin),
			},
			Title: node.Title,
		}

		i++
	}
}
