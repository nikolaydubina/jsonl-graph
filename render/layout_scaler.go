package render

// ScalerLayout will scale existing layout by constant factor.
type ScalerLayout struct {
	Scale float64
}

func (l *ScalerLayout) UpdateGraphLayout(g Graph) {
	for i := range g.Nodes {
		x := float64(g.Nodes[i].LeftBottom.X)
		y := float64(g.Nodes[i].LeftBottom.Y)

		g.Nodes[i].LeftBottom.X = int(x * l.Scale)
		g.Nodes[i].LeftBottom.Y = int(y * l.Scale)
	}

	// TODO: adjust that node center is still fixed size
	for e := range g.Edges {
		for i := range g.Edges[e].Points {
			x := float64(g.Edges[e].Points[i].X)
			y := float64(g.Edges[e].Points[i].Y)
			g.Edges[e].Points[i].X = int(x * l.Scale)
			g.Edges[e].Points[i].Y = int(y * l.Scale)
		}
	}
}
