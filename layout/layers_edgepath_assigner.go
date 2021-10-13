package layout

type StraightEdgePathAssigner struct {
	MarginY        int
	MarginX        int
	FakeNodeWidth  int
	FakeNodeHeight int
}

func (l StraightEdgePathAssigner) UpdateGraphLayout(g Graph, lg LayeredGraph) {
	yOffset := 0
	for _, nodes := range lg.Layers() {
		xOffset := 0
		hMax := 0

		for _, node := range nodes {
			w := 0
			h := 0

			if lg.Dummy[node] {
				w = l.FakeNodeWidth
				h = l.FakeNodeHeight
			} else {
				w = g.Nodes[node].W
				h = g.Nodes[node].H
			}

			g.Nodes[node] = Node{
				XY: [2]int{lg.NodeYX[node][1], yOffset},
				W:  g.Nodes[node].W,
				H:  g.Nodes[node].H,
			}

			xOffset += w
			if h > hMax {
				hMax = h
			}
		}

		yOffset += hMax + l.MarginY
	}

	DirectEdgesLayout{}.UpdateGraphLayout(g)

	for node := range lg.Dummy {
		delete(g.Nodes, node)
	}
}
