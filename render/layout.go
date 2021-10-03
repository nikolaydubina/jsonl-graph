package render

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

// MemoLayout computes layout for memoized and stores to target.
type MemoLayout struct {
	Graph  Graph
	Layout Layout
}

func (l MemoLayout) UpdateGraphLayout(g Graph) {
	newgraph := l.Graph.Copy()
	l.Layout.UpdateGraphLayout(newgraph)

	for i := range g.Nodes {
		g.Nodes[i].LeftBottom = newgraph.Nodes[i].LeftBottom
	}
	for e := range g.Edges {
		edge := *newgraph.Edges[e]
		g.Edges[e] = &edge
	}
}
