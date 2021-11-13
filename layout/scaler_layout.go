package layout

// ScalerLayout will scale existing layout by constant factor for nodes and recompute edges layout.
type ScalerLayout struct {
	Scale      float64
	EdgeLayout Layout
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

	l.EdgeLayout.UpdateGraphLayout(g)
}
