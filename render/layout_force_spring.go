package render

import "math"

// SpringForce is linear by distance
type SpringForce struct {
	K         float64 // has to be positive
	L         float64 // distance at rest
	EdgesOnly bool    // true = only edges, false = all nodes
}

func (l SpringForce) AddForce(g Graph, fx map[uint64]float64, fy map[uint64]float64) {
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
				// if stretch, then attraction
				// if shrink, then repulsion
				f := (d - l.L) * l.K
				fx[i] += f * (xj - xi) / d
				fy[i] += f * (yj - yi) / d
			}
		}
	}
}
