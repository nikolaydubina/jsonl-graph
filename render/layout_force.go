package render

import "math"

// Force computes forces for Nodes.
type Force interface {
	AddForce(g Graph, fx map[uint64]float64, fy map[uint64]float64)
}

// ForceGraphLayout will simulate node movement due to forces.
type ForceGraphLayout struct {
	Delta    float64 // how much move each step
	MaxSteps int     // limit of iterations
	Epsilon  float64 // minimal force
	Forces   []Force
}

func (l ForceGraphLayout) UpdateGraphLayout(g Graph) {
	for step := 0; step < l.MaxSteps; step++ {
		fx := make(map[uint64]float64, len(g.Nodes))
		fy := make(map[uint64]float64, len(g.Nodes))

		// accumulate all forces
		for i := range l.Forces {
			l.Forces[i].AddForce(g, fx, fy)
		}

		// delete tiny forces
		for i := range g.Nodes {
			if math.Hypot(fx[i], fy[i]) < l.Epsilon {
				delete(fx, i)
				delete(fy, i)
			}
		}

		// early stop if no forces
		if len(fx) == 0 || len(fy) == 0 {
			break
		}

		// move by delta
		for i := range g.Nodes {
			g.Nodes[i].LeftBottom.X += int(fx[i] * l.Delta)
			g.Nodes[i].LeftBottom.Y += int(fy[i] * l.Delta)
		}

	}

	for e := range g.Edges {
		edge := DirectEdge(*g.Nodes[e[0]], *g.Nodes[e[1]])
		g.Edges[e] = &edge
	}
}
