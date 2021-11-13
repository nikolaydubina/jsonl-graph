package layout

// Expects that graph g does not have cycles.
func NewLayeredGraph(g Graph) LayeredGraph {
	nodeYX := make(map[uint64][2]int, len(g.Nodes))

	for _, root := range g.Roots() {
		nodeYX[root] = [2]int{0, 0}
		for que := []uint64{root}; len(que) > 0; {
			// pop
			p := que[0]
			if len(que) > 1 {
				que = que[1:]
			} else {
				que = nil
			}

			// set max depth for each child
			for e := range g.Edges {
				if parent, child := e[0], e[1]; parent == p {
					if l := nodeYX[parent][0] + 1; l > nodeYX[child][0] {
						nodeYX[child] = [2]int{l, 0}
					}
					que = append(que, child)
				}
			}
		}
	}

	// segments
	segments := map[[2]uint64]bool{}
	for e := range g.Edges {
		segments[e] = true
	}

	return LayeredGraph{
		NodeYX:   nodeYX,
		Segments: segments,
		Dummy:    map[uint64]bool{},
	}
}
