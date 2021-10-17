package layout

type Layout interface {
	UpdateGraphLayout(g Graph)
}

// CompositeLayout will apply multiple layouts in sequence.
type CompositeLayout struct {
	Layouts []Layout
}

func (l CompositeLayout) UpdateGraphLayout(g Graph) {
	for _, layout := range l.Layouts {
		layout.UpdateGraphLayout(g)
	}
}

// MemoLayout applies layout updaters to memoized graph and stores to destination.
type MemoLayout struct {
	Graph  Graph
	Layout Layout
}

func (l MemoLayout) UpdateGraphLayout(g Graph) {
	newgraph := CopyGraph(l.Graph)

	l.Layout.UpdateGraphLayout(newgraph)

	// apply to target graph
	for i := range g.Nodes {
		g.Nodes[i] = Node{
			XY: newgraph.Nodes[i].XY,
			W:  g.Nodes[i].W,
			H:  g.Nodes[i].H,
		}
	}
	for e := range g.Edges {
		g.Edges[e] = Edge{Path: make([][2]int, len(newgraph.Edges[e].Path))}
		for i, ne := range newgraph.Edges[e].Path {
			g.Edges[e].Path[i] = ne
		}
	}
}
