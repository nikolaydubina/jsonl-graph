package render

import "math"

// GravityEdgesForce is gravity attraction force by edges: f_ij = k * / (x_j - x_i) ** 2)
type GravityEdgesForce struct {
	K float64 // how to translate distance to force
}

func (l GravityEdgesForce) Force(g Graph) map[uint64][2]float64 {
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

				if d > 1 {
					f := l.K / (d * d)
					fx += (xj - xi) / d * f
					fy += (yj - yi) / d * f
				}
			}
		}

		forces[i] = [2]float64{fx, fy}
	}

	return forces
}

// GravityNodesForce is gravity attraction force among all nodes: f_ij = k * / (x_j - x_i) ** 2)
type GravityNodesForce struct {
	K float64 // how to translate distance to force
}

func (l GravityNodesForce) Force(g Graph) map[uint64][2]float64 {
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

			if d > 1 {
				f := l.K / (d * d)
				fx += (xj - xi) / d * f
				fy += (yj - yi) / d * f
			}
		}

		forces[i] = [2]float64{fx, fy}
	}

	return forces
}
