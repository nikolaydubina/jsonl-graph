package layout

type CycleRemover interface {
	RemoveCycles(g Graph)
	Restore(g Graph)
}

// Kozo Sugiyama algorithm breaks down layered graph construction in phases.
type SugiyamaLayersStrategyGraphLayout struct {
	CycleRemover     CycleRemover
	LevelsAssigner   func(g Graph) LayeredGraph
	OrderingAssigner func(g Graph, lg LayeredGraph)
	XAssigner        func(g Graph, lg LayeredGraph)
	EdgePathAssigner func(g Graph, lg LayeredGraph)
}

func (l SugiyamaLayersStrategyGraphLayout) UpdateGraphLayout(g Graph) {
	l.CycleRemover.RemoveCycles(g)
	lg := l.LevelsAssigner(g)
	l.OrderingAssigner(g, lg)
	l.XAssigner(g, lg)
	l.EdgePathAssigner(g, lg)
	l.CycleRemover.Restore(g)
}

// TODO: resolve vertical coordinate
func NewBasicSugiyamaLayersGraphLayout() SugiyamaLayersStrategyGraphLayout {
	return SugiyamaLayersStrategyGraphLayout{
		CycleRemover:   NewSimpleCycleRemover(),
		LevelsAssigner: NewLayeredGraph,
		OrderingAssigner: LBLOrderingOptimizer{
			Epochs: 10,
			LayerOrderingOptimizer: RandomLayerOrderingOptimizer{
				Epochs: 5,
			},
		}.Optimize,
		XAssigner: BrandesKopfHorizontalAssigner{
			Delta: 25, // TODO: dynamically from graph width
		}.AssignX,
		EdgePathAssigner: StraightEdgePathAssigner{
			MarginY:        25,
			MarginX:        25,
			FakeNodeWidth:  25,
			FakeNodeHeight: 25,
		}.UpdateGraphLayout,
	}
}
