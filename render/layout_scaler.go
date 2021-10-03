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

	for e := range g.Edges {
		edge := DirectEdge(*g.Nodes[e[0]], *g.Nodes[e[1]])
		g.Edges[e] = &edge
	}
}
