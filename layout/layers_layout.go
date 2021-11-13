package layout

type CycleRemover interface {
	RemoveCycles(g Graph)
	Restore(g Graph)
}

type NodesHorizontalCoordinatesAssigner interface {
	NodesHorizontalCoordinates(g LayeredGraph) map[uint64]int
}

// Kozo Sugiyama algorithm breaks down layered graph construction in phases.
type SugiyamaLayersStrategyGraphLayout struct {
	CycleRemover                       CycleRemover
	LevelsAssigner                     func(g Graph) LayeredGraph
	OrderingAssigner                   func(g Graph, lg LayeredGraph)
	NodesHorizontalCoordinatesAssigner NodesHorizontalCoordinatesAssigner
	EdgePathAssigner                   func(g Graph, lg LayeredGraph)
}

// UpdateGraphLayout breaks down layered graph construction in phases.
func (l SugiyamaLayersStrategyGraphLayout) UpdateGraphLayout(g Graph) {
	l.CycleRemover.RemoveCycles(g)

	// assign levels to graph
	lg := l.LevelsAssigner(g)
	if err := lg.Validate(); err != nil {
		panic(err)
	}

	// assign order withing levels
	l.OrderingAssigner(g, lg)

	// calculate nodes horizontal coordinates
	nodeX := l.NodesHorizontalCoordinatesAssigner.NodesHorizontalCoordinates(lg)
	for node, x := range nodeX {
		yx := lg.NodeYX[node]
		yx[1] = x
		lg.NodeYX[node] = yx
	}

	// TODO: resolve vertical coordinate

	l.EdgePathAssigner(g, lg)
	l.CycleRemover.Restore(g)
}
