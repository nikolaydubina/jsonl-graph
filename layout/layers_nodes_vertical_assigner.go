package layout

// BasicNodesVerticalCoordinatesAssigner will check maximum height in each layer.
// It will keep each node vertically in the middle within each layer.
type BasicNodesVerticalCoordinatesAssigner struct {
	MarginLayers   int // distance between layers
	FakeNodeHeight int
}

func layersMaxHeights(g Graph, lg LayeredGraph) []int {
	hmax := make([]int, len(lg.Layers()))

	for i, nodes := range lg.Layers() {
		for _, node := range nodes {
			if hmax[i] < g.Nodes[node].H {
				hmax[i] = g.Nodes[node].H
			}
		}
	}

	return hmax
}

func (s BasicNodesVerticalCoordinatesAssigner) NodesVerticalCoordinates(g Graph, lg LayeredGraph) map[uint64]int {
	nodeY := make(map[uint64]int, len(lg.NodeYX))

	layersHMax := layersMaxHeights(g, lg)

	yOffset := 0
	for i, nodes := range lg.Layers() {
		for _, node := range nodes {
			nodeH := s.FakeNodeHeight
			if n, ok := g.Nodes[node]; ok {
				// if not fake node, then set its actual height
				nodeH = n.H
			}

			// put in the middle vertically
			nodeY[node] = yOffset + ((layersHMax[i] - nodeH) / 2)
		}

		// move to next layer
		yOffset += layersHMax[i] + s.MarginLayers
	}

	return nodeY
}
