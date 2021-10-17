package layout

type LevelsAssigner func(g Graph) LayeredGraph

// Kozo Sugiyama algorithm breaks down layered graph construction in phases.
// Expects directed acyclic graph.
// TODO: add preprocessing for un-directed/cyclic graphs.
type SugiyamaLayersStrategyGraphLayout struct {
	LevelsAssigner   func(g Graph) LayeredGraph
	OrderingAssigner func(g Graph, lg LayeredGraph)
	XAssigner        func(g Graph, lg LayeredGraph)
	EdgePathAssigner func(g Graph, lg LayeredGraph)
}

func (l SugiyamaLayersStrategyGraphLayout) UpdateGraphLayout(g Graph) {
	lg := l.LevelsAssigner(g)
	l.OrderingAssigner(g, lg)
	l.XAssigner(g, lg)
	l.EdgePathAssigner(g, lg)
}

func NewBasicSugiyamaLayersGraphLayout() SugiyamaLayersStrategyGraphLayout {
	return SugiyamaLayersStrategyGraphLayout{
		LevelsAssigner: NewLayeredGraph,
		OrderingAssigner: LBLOrderingOptimizer{
			Epochs: 10,
			LayerOrderingOptimizer: RandomLayerOrderingOptimizer{
				Epochs: 5,
			},
		}.Optimize,
		XAssigner: BrandesKopfHorizontalAssigner{
			Delta: 25,
		}.AssignX,
		EdgePathAssigner: StraightEdgePathAssigner{
			MarginY:        25,
			MarginX:        25,
			FakeNodeWidth:  25,
			FakeNodeHeight: 25,
		}.UpdateGraphLayout,
	}
}
