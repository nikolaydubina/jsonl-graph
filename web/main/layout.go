package app

import (
	"log"
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
		switch layoutOption {
		case ForcesLayoutOption:
			r.layoutUpdater = layout.ForceGraphLayout{
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
			r.layoutUpdater = layout.EadesGonumLayout{
				Repulsion: 1,
				Rate:      0.05,
				Updates:   30,
				Theta:     0.2,
				ScaleX:    0.5,
				ScaleY:    0.5,
			}
		case IsomapLayoutOption:
			r.layoutUpdater = layout.IsomapR2GonumLayout{
				ScaleX: 0.5,
				ScaleY: 0.5,
			}
		case LayersLayoutOption:
			r.layoutUpdater = layout.NewBasicSugiyamaLayersGraphLayout()
		default:
			log.Printf("unexpected layout option(%s)", layoutOption)
		}

		r.SetInitialUpdateGraphLayout()
		r.Render()
		return nil
	}
}

// SetInitialUpdateGraphLayout sets layout to what it should look like at the begging for a layout.
func (r *Bridge) SetInitialUpdateGraphLayout() {
	r.layoutUpdater.UpdateGraphLayout(r.graphLayout)
	r.scalerLayout.UpdateGraphLayout(r.graphLayout)

	r.scalerLayout.Graph = layout.CopyGraph(r.graphLayout)
	CenterGraph(r.graphRender, r.scaler)
}
