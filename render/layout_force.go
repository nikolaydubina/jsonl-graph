package render

// Force computes forces for Nodes.
type Force interface {
	Force(g Graph) map[uint64][2]float64
}

// ForceGraphLayout will simulate node movement due to forces.
// TODO: add early stop if nodes did not move within some constant.
type ForceGraphLayout struct {
	Delta    float64 // how much move each step
	MaxSteps int     // limit of iterations
	Forces   []Force
}

func (l ForceGraphLayout) UpdateGraphLayout(g Graph) {
	for step := 0; step < l.MaxSteps; step++ {
		forces := make(map[uint64][2]float64, len(g.Nodes))

		// accumulate all forces
		for i := range l.Forces {
			for k, v := range l.Forces[i].Force(g) {
				forces[k] = [2]float64{forces[k][0] + v[0], forces[k][1] + v[1]}
			}
		}

		// move by delta
		for i := range g.Nodes {
			g.Nodes[i].LeftBottom.X += int(forces[i][0] * l.Delta)
			g.Nodes[i].LeftBottom.Y += int(forces[i][1] * l.Delta)
		}

	}

	// update edges
	for idFrom, toEdges := range g.Edges {
		for idTo := range toEdges {
			edge := DirectEdge(*g.Nodes[idFrom], *g.Nodes[idTo])
			g.Edges[idFrom][idTo] = &edge
		}
	}

}
