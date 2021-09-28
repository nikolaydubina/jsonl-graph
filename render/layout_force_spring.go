package render

import "math"

// ShrinkSpringEdgesForce is pushing away nodes if they try to shrink bellow threshold for edges: f_ij = k * (l - (x_j - x_i))
type ShrinkSpringEdgesForce struct {
	K float64 // how to translate distance to force
	L float64 // distance at rest
}

func (l ShrinkSpringEdgesForce) Force(g Graph) map[uint64][2]float64 {
	forces := make(map[uint64][2]float64, len(g.Nodes))

	// force applied to node i
	for i := range g.Nodes {
		fx := 0.0
		fy := 0.0

		xi := float64(g.Nodes[i].LeftBottom.X)
		yi := float64(g.Nodes[i].LeftBottom.Y)

		if toIDs, ok := g.Edges[i]; ok {
			for toID := range toIDs {
				nodej := g.Nodes[toID]

				xj := float64(nodej.LeftBottom.X)
				yj := float64(nodej.LeftBottom.Y)

				d := math.Hypot(xi-xj, yi-yj)

				if d < l.L && d > 1 {
					f := (l.L - d) * l.K
					fx = -1 * (xj - xi) / d * f
					fy = -1 * (yj - yi) / d * f
				}
			}
		}

		forces[i] = [2]float64{fx, fy}
	}

	return forces
}

// ShrinkSpringNodesForce is pushing away nodes if they try to shrink bellow threshold for all nodes: f_ij = k * (l - (x_j - x_i))
type ShrinkSpringNodesForce struct {
	K float64 // how to translate distance to force
	L float64 // distance at rest
}

func (l ShrinkSpringNodesForce) Force(g Graph) map[uint64][2]float64 {
	forces := make(map[uint64][2]float64, len(g.Nodes))

	// force applied to node i
	for i := range g.Nodes {
		fx := 0.0
		fy := 0.0

		xi := float64(g.Nodes[i].LeftBottom.X)
		yi := float64(g.Nodes[i].LeftBottom.Y)

		for j := range g.Nodes {
			if i == j {
				continue
			}

			xj := float64(g.Nodes[j].LeftBottom.X)
			yj := float64(g.Nodes[j].LeftBottom.Y)

			d := math.Hypot(xi-xj, yi-yj)

			if d < l.L && d > 1 {
				f := (l.L - d) * l.K
				fx = -1 * (xj - xi) / d * f
				fy = -1 * (yj - yi) / d * f
			}
		}

		forces[i] = [2]float64{fx, fy}
	}

	return forces
}
