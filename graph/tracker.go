package graph

import "log"

// GraphTracker tells if graph has changed structurally.
// Not considering contents of nodes or edges.
// Only considering connection between nodes, number of nodes, and direction of edges.
type GraphTracker struct {
	nodes map[uint64]bool
	edges map[[2]uint64]bool
}

func NewGraphTracker(g Graph) GraphTracker {
	nodes := make(map[uint64]bool, len(g.Nodes))
	for id := range g.Nodes {
		nodes[id] = true
	}

	edges := make(map[[2]uint64]bool, len(g.Edges))
	for e := range g.Edges {
		edges[e] = true
	}

	return GraphTracker{
		nodes: nodes,
		edges: edges,
	}
}

func (og GraphTracker) HasChanged(g Graph) bool {
	if len(g.Nodes) != len(og.nodes) {
		log.Printf("graph tracker: num nodes canged from(%d) to(%d)", len(og.nodes), len(g.Nodes))
		return true
	}

	if len(g.Edges) != len(og.edges) {
		log.Printf("graph tracker: num edges canged from(%d) to(%d)", len(og.edges), len(g.Edges))
		return true
	}

	for id := range g.Nodes {
		if !og.nodes[id] {
			log.Printf("graph tracker: new node not found(%d)", id)
			return true
		}
	}

	for e := range g.Edges {
		if !og.edges[e] {
			log.Printf("graph tracker: new edge not found(%v)", e)
			return true
		}
	}
	return false
}
