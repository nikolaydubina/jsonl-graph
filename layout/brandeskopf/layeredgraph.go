package brandeskopf

// LayeredGraph is graph with Dummy nodes such that there is no long edges.
// Short edge is between nodes in Layers next to each other.
// Long edge is between nodes in 1+ Layers between each other.
// Segments are short edges and long edges.
// Top layer has lowest layer number.
type LayeredGraph struct {
	Segments map[[2]uint64]bool // segment is an edge in layered graph, can be real edge or piece of fake edge
	Dummy    map[uint64]bool    // fake nodes
	NodeYX   map[uint64][2]int  // node -> {layer, ordering in layer}
	Layers   [][]uint64         // same as NodeYX but different form
}

// IsInnerSegment tells when edge is between two Dummy nodes.
func (g LayeredGraph) IsInnerSegment(segment [2]uint64) bool {
	return g.Dummy[segment[0]] && g.Dummy[segment[1]]
}

// UpperNeighbors are nodes in upper layer that are connected to given node.
func (g LayeredGraph) UpperNeighbors(node uint64) []uint64 {
	var nodes []uint64
	for e := range g.Segments {
		if e[1] == node {
			if g.NodeYX[e[0]][0] == (g.NodeYX[e[1]][0] - 1) {
				nodes = append(nodes, e[0])
			}
		}
	}
	return nodes
}

// LowerNeighbors are nodes in lower layer that are connected to given node.
func (g LayeredGraph) LowerNeighbors(node uint64) []uint64 {
	var nodes []uint64
	for e := range g.Segments {
		if e[0] == node {
			if g.NodeYX[e[0]][0] == (g.NodeYX[e[1]][0] - 1) {
				nodes = append(nodes, e[0])
			}
		}
	}
	return nodes
}
