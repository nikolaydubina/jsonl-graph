package layout

// ScalerLayout will scale existing layout by constant factor.
type ScalerLayout struct {
	Scale float64
}

func (l *ScalerLayout) UpdateGraphLayout(g Graph) {
	for i := range g.Nodes {
		x := float64(g.Nodes[i].XY[0])
		y := float64(g.Nodes[i].XY[1])

		g.Nodes[i] = Node{
			XY: [2]int{int(x * l.Scale), int(y * l.Scale)},
			W:  g.Nodes[i].W,
			H:  g.Nodes[i].H,
		}
	}

	// TODO: adjust that node center is still fixed size
	for e := range g.Edges {
		for i := range g.Edges[e].Path {
			x := float64(g.Edges[e].Path[i][0])
			y := float64(g.Edges[e].Path[i][1])
			g.Edges[e].Path[i] = [2]int{int(x * l.Scale), int(y * l.Scale)}
		}
	}
}
