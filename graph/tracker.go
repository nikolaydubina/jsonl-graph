package graph

import "log"

// GraphTracker tells if graph has changed structurally.
// Not considering contents of nodes or edges.
// Only considering connection between nodes, number of nodes, and direction of edges.
type GraphTracker struct {
	oldNodes map[uint64]bool
	oldEdges map[[2]uint64]bool
}

func NewGraphTracker(g Graph) GraphTracker {
	oldNodes := make(map[uint64]bool, len(g.Nodes))
	for id := range g.Nodes {
		oldNodes[id] = true
	}

	oldEdges := make(map[[2]uint64]bool)
	for from, tos := range g.Edges {
		for to := range tos {
			oldEdges[[2]uint64{from, to}] = true
		}
	}
	return GraphTracker{
		oldNodes: oldNodes,
		oldEdges: oldEdges,
	}
}

func (og GraphTracker) HasChanged(g Graph) bool {
	if len(g.Nodes) != len(og.oldNodes) {
		log.Printf("num nodes canged from(%d) to(%d)", len(g.Nodes), len(og.oldNodes))
		return true
	}
	for id := range g.Nodes {
		if !og.oldNodes[id] {
			log.Printf("new node not found(%d)", id)
			return true
		}
	}

	numEdges := 0
	for _, tos := range g.Edges {
		numEdges += len(tos)
	}
	if numEdges != len(og.oldEdges) {
		log.Printf("num edges canged from(%d) to(%d)", numEdges, len(og.oldEdges))
		return true
	}
	for from, tos := range g.Edges {
		for to := range tos {
			if !og.oldEdges[[2]uint64{from, to}] {
				log.Printf("new edge not found(%d, %d)", from, to)
				return true
			}
		}
	}
	return false
}
