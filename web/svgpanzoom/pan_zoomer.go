package svgpanzoom

import (
	"fmt"
	"math"
	"strings"
	"syscall/js"

	"gonum.org/v1/gonum/mat"
)

type State string

const (
	Drag      State = "drag"
	NotActive State = ""
)

// PanZoomer handles updates to SVG based
// This is port of https://github.com/aleofreddi/svgpan
// Which is also as of 2021-09-19 inlined in https://github.com/google/pprof/blob/master/third_party/svgpan/svgpan.go
type PanZoomer struct {
	rootID    string
	zoomScale float64
	state     State
	transform *mat.Dense
	origin    *mat.Dense
}

func NewPanZoomer(
	rootID string,
	zoomScale float64,
) *PanZoomer {
	return &PanZoomer{
		rootID:    rootID,
		zoomScale: zoomScale,
		transform: newIdentity(),
	}
}

func (p *PanZoomer) SetupHandlers() {
	js.Global().Set("handleMouseUp", js.FuncOf(p.handleMouseUp))
	js.Global().Set("handleMouseDown", js.FuncOf(p.handleMouseDown))
	js.Global().Set("handleMouseMove", js.FuncOf(p.handleMouseMove))

	js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "mouseup", js.FuncOf(p.handleMouseUp))
	js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "mousedown", js.FuncOf(p.handleMouseDown))
	js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "mousemove", js.FuncOf(p.handleMouseMove))
	js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "mousepout", js.FuncOf(p.handleMouseUp))

	userAgent := js.Global().Get("navigator").Get("userAgent").String()
	if strings.Contains(strings.ToLower(userAgent), "webkit") {
		// Chrome/Safari
		js.Global().Get("window").Call("addEventListener", "mousewheel", js.FuncOf(p.handleMouseWheel), false)
	} else {
		// Firefox/Other
		js.Global().Get("window").Call("addEventListener", "DOMMouseScroll", js.FuncOf(p.handleMouseWheel), false)
	}
}

func getEventPoint(event js.Value) *mat.Dense {
	return mat.NewDense(4, 1, []float64{
		event.Get("clientX").Float(),
		event.Get("clientY").Float(),
		0,
		1,
	})
}

// setRootTranslation updates translation matrix of root svg element
func (p *PanZoomer) setRootTranslation() {
	m := p.transform

	a := m.At(0, 0)
	b := m.At(0, 1)
	c := m.At(1, 0)
	d := m.At(1, 1)
	e := m.At(2, 0)
	f := m.At(2, 1)

	s := fmt.Sprintf("matrix(%f,%f,%f,%f,%f,%f)", a, b, c, d, e, f)
	js.Global().Get("document").Call("getElementById", p.rootID).Call("setAttribute", "transform", s)
}

func (p *PanZoomer) handleMouseWheel(this js.Value, args []js.Value) interface{} {
	event := args[0]
	delta := 0.0
	if event.Get("wheelDelta").Truthy() {
		// Chrome/Safari
		delta = event.Get("wheelDelta").Float() / 360
	} else {
		// Mozilla
		delta = event.Get("detail").Float() / -9
	}

	var z = math.Pow(1+p.zoomScale, delta)

	var point mat.Dense
	point.Mul(p.transform, getEventPoint(event))

	// Compute new scale matrix in current mouse position
	x := point.At(0, 0)
	y := point.At(1, 0)

	var k *mat.Dense = newIdentity()
	k.Mul(newTranslate(x, y, 0), k)
	k.Mul(newScale(z), k)
	k.Mul(newTranslate(-x, -y, 0), k)
	k.Inverse(k)

	p.transform.Mul(p.transform, k)
	p.setRootTranslation()
	return nil
}

func (p *PanZoomer) handleMouseMove(_ js.Value, args []js.Value) interface{} {
	if p.state != Drag {
		return nil
	}

	event := args[0]
	var point mat.Dense
	point.Mul(p.transform, getEventPoint(event))

	x := point.At(0, 0)
	y := point.At(1, 0)
	ox := p.origin.At(0, 0)
	oy := p.origin.At(1, 0)

	var k mat.Dense
	k.Inverse(p.transform)
	k.Mul(p.transform, newTranslate(x-ox, y-oy, 0))
	k.Inverse(&k)

	p.transform.Mul(p.transform, &k)
	p.setRootTranslation()
	return nil
}

func (p *PanZoomer) handleMouseDown(_ js.Value, args []js.Value) interface{} {
	p.state = Drag
	event := args[0]
	p.origin = getEventPoint(event)
	return nil
}

func (p *PanZoomer) handleMouseUp(_ js.Value, _ []js.Value) interface{} {
	p.state = NotActive
	return nil
}

func newIdentity() *mat.Dense {
	return mat.NewDense(4, 4, []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	})
}

func newScale(z float64) *mat.Dense {
	return mat.NewDense(4, 4, []float64{
		z, 0, 0, 0,
		0, z, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	})
}

func newTranslate(x, y, z float64) *mat.Dense {
	return mat.NewDense(4, 4, []float64{
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	})
}
