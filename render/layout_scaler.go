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

	//  update edges
	for idFrom, toEdges := range g.Edges {
		for idTo := range toEdges {
			edge := DirectEdge(*g.Nodes[idFrom], *g.Nodes[idTo])
			g.Edges[idFrom][idTo] = &edge
		}
	}
}
