package renderer

import (
	"syscall/js"

	"github.com/nikolaydubina/jsonl-graph/render"
)

type LayoutOption string

const (
	GridLayoutOption   LayoutOption = "layoutOptionGrid"
	ForcesLayoutOption LayoutOption = "layoutOptionForces"
	EadesLayoutOption  LayoutOption = "layoutOptionEades"
	IsomapLayoutOption LayoutOption = "layoutOptionIsomap"
)

// NewLayoutOptionUpdater constructs new handler for layout option.
// TODO: read options of layout from UI
func (r *Renderer) NewLayoutOptionUpdater(layoutOption LayoutOption) func(_ js.Value, _ []js.Value) interface{} {
	return func(_ js.Value, _ []js.Value) interface{} {
		switch layoutOption {
		case GridLayoutOption:
			r.layoutUpdater = render.BasicGridLayout{
				RowLength: 5,
				Margin:    25,
			}
		case ForcesLayoutOption:
			render.InitRandom(r.graphRender)
			r.layoutUpdater = render.ForceGraphLayout{
				Delta:    1,
				MaxSteps: 5000,
				Epsilon:  1.5,
				Forces: []render.Force{
					render.GravityForce{
						K:         -50,
						EdgesOnly: false,
					},
					render.SpringForce{
						K:         0.2,
						L:         200,
						EdgesOnly: true,
					},
				},
			}
		case EadesLayoutOption:
			r.layoutUpdater = render.EadesGonumLayout{
				Repulsion: 1,
				Rate:      0.05,
				Updates:   30,
				Theta:     0.2,
				ScaleX:    0.5,
				ScaleY:    0.5,
			}
		case IsomapLayoutOption:
			r.layoutUpdater = render.IsomapR2GonumLayout{
				ScaleX: 0.5,
				ScaleY: 0.5,
			}
		}

		r.layoutUpdater.UpdateGraphLayout(r.graphRender)
		CenterGraph(r.graphRender, r.scaler)
		r.Render()
		return nil
	}
}
