package layout

// Layout is something that can update graph layout
type Layout interface {
	UpdateGraphLayout(g Graph)
}
