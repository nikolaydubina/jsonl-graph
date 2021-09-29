package render

import "math"

// GravityForce is gravity force
type GravityForce struct {
	K         float64 // positive K for attraction
	EdgesOnly bool    // true = only edges, false = all nodes
}

func (l GravityForce) AddForce(g Graph, fx map[uint64]float64, fy map[uint64]float64) {
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

		xi := float64(g.Nodes[i].LeftBottom.X)
		yi := float64(g.Nodes[i].LeftBottom.Y)

		for _, j := range js {
			xj := float64(g.Nodes[j].LeftBottom.X)
			yj := float64(g.Nodes[j].LeftBottom.Y)

			d := math.Hypot(xi-xj, yi-yj)

			if d > 1 {
				f := l.K / d
				fx[i] += f * (xj - xi) / d
				fy[i] += f * (yj - yi) / d
			}
		}
	}
}
