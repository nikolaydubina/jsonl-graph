package render

import "math"

// GravityForce is gravity force
type GravityForce struct {
	K         float64 // positive K for attraction
	EdgesOnly bool    // true = only edges, false = all nodes
}

func (l GravityForce) Force(g Graph) map[uint64][2]float64 {
	forces := make(map[uint64][2]float64, len(g.Nodes))

	// force applied to node i
	for i := range g.Nodes {
		var js []uint64
		if l.EdgesOnly {
			if toIDs, ok := g.Edges[i]; ok {
				for j := range toIDs {
					js = append(js, j)
				}
			}
		} else {
			for j := range g.Nodes {
				if i != j {
					js = append(js, j)
				}
			}
		}

		fx := 0.0
		fy := 0.0

		xi := float64(g.Nodes[i].LeftBottom.X)
		yi := float64(g.Nodes[i].LeftBottom.Y)

		for _, j := range js {
			xj := float64(g.Nodes[j].LeftBottom.X)
			yj := float64(g.Nodes[j].LeftBottom.Y)

			d := math.Hypot(xi-xj, yi-yj)

			if d > 1 {
				f := l.K / d
				fx += f * (xj - xi) / d
				fy += f * (yj - yi) / d
			}
		}

		forces[i] = [2]float64{fx, fy}
	}

	return forces
}
