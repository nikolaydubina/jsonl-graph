package app

import (
	"syscall/js"

	"github.com/nikolaydubina/jsonl-graph/layout"
)

type LayoutOption string

const (
	ForcesLayoutOption LayoutOption = "layoutOptionForces"
	EadesLayoutOption  LayoutOption = "layoutOptionEades"
	IsomapLayoutOption LayoutOption = "layoutOptionIsomap"
	LayersLayoutOption LayoutOption = "layoutOptionLayers"
)

func AllLayoutOptions() []LayoutOption {
	return []LayoutOption{
		ForcesLayoutOption,
		EadesLayoutOption,
		IsomapLayoutOption,
		LayersLayoutOption,
	}
}

// NewLayoutOptionUpdater constructs new handler for layout option.
// TODO: read options of layout from UI
func (r *Bridge) NewLayoutOptionUpdater(layoutOption LayoutOption) func(_ js.Value, _ []js.Value) interface{} {
	return func(_ js.Value, _ []js.Value) interface{} {
		var mainLayout layout.Layout

		switch layoutOption {
		case ForcesLayoutOption:
			layout.InitRandom(r.graphRender)
			mainLayout = layout.ForceGraphLayout{
				Delta:    1,
				MaxSteps: 5000,
				Epsilon:  1.5,
				Forces: []layout.Force{
					layout.GravityForce{
						K:         -50,
						EdgesOnly: false,
					},
					layout.SpringForce{
						K:         0.2,
						L:         200,
						EdgesOnly: true,
					},
				},
			}
		case EadesLayoutOption:
			mainLayout = layout.EadesGonumLayout{
				Repulsion: 1,
				Rate:      0.05,
				Updates:   30,
				Theta:     0.2,
				ScaleX:    0.5,
				ScaleY:    0.5,
			}
		case IsomapLayoutOption:
			mainLayout = layout.IsomapR2GonumLayout{
				ScaleX: 0.5,
				ScaleY: 0.5,
			}
		case LayersLayoutOption:
			mainLayout = layout.NewBasicSugiyamaLayersGraphLayout()
		}

		r.layoutUpdater = layout.CompositeLayout{
			Layouts: []layout.Layout{
				mainLayout,
				r.scalerLayout.Layout, // inner layout without memoization
			},
		}

		r.layoutUpdater.UpdateGraphLayout(r.graphRender)
		CenterGraph(r.graphRender, r.scaler)

		// update memoized graph for scaling
		r.scalerLayout.Graph = r.graphRender.Copy()

		r.Render()
		return nil
	}
}
