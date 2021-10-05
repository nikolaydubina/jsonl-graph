package app

import (
	"log"
	"syscall/js"

	"github.com/nikolaydubina/jsonl-graph/render"
	"github.com/nikolaydubina/jsonl-graph/web/svgpanzoom"
)

// CenterGraph will update scaler with transformations such that
// given graph will be in the center and fills the screen.
func CenterGraph(g render.Graph, scaler *svgpanzoom.PanZoomer) {
	minx, miny, maxx, maxy := g.BoundingBox()
	wScreen := js.Global().Get("window").Get("innerWidth").Float()
	hScreen := js.Global().Get("window").Get("innerHeight").Float()

	dx, dy, zoom := render.CenterBox(wScreen, hScreen, float64(minx), float64(miny), float64(maxx), float64(maxy))
	log.Printf("centering: screen (%f x %f) graph (%d %d %d %d) thus transform dx(%f) dy(%f) zoom(%f)", wScreen, hScreen, minx, miny, maxx, maxy, dx, dy, zoom)
	scaler.Reset().Translate(dx, dy).ScaleAt(zoom, wScreen/2, hScreen/2)
}
