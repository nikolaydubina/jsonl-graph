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
	rootID              string
	zoomScale           float64
	state               State
	transform           *mat.Dense
	transformBeforeDrag *mat.Dense
	origin              *mat.Dense
}

func NewPanZoomer(
	rootID string,
	zoomScale float64,
) *PanZoomer {
	return &PanZoomer{
		rootID:    rootID,
		zoomScale: zoomScale,
		transform: identity(),
	}
}

func (p *PanZoomer) SetupHandlers() {
	js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "mouseup", js.FuncOf(p.handleMouseUp))
	js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "mousedown", js.FuncOf(p.handleMouseDown))
	js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "mousemove", js.FuncOf(p.handleMouseMove))
	js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "mousepout", js.FuncOf(p.handleMouseUp))

	userAgent := js.Global().Get("navigator").Get("userAgent").String()
	if strings.Contains(strings.ToLower(userAgent), "webkit") {
		// Chrome/Safari
		js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "mousewheel", js.FuncOf(p.handleMouseWheel), false)
	} else {
		// Firefox/Other
		js.Global().Get("document").Call("getElementById", "graph2").Call("addEventListener", "DOMMouseScroll", js.FuncOf(p.handleMouseWheel), false)
	}
}

func getEventPoint(event js.Value) *mat.Dense {
	return mat.NewDense(3, 1, []float64{
		event.Get("clientX").Float(),
		event.Get("clientY").Float(),
		1,
	})
}

// setRootTranslation updates translation matrix of root svg element
func (p *PanZoomer) setRootTranslation() {
	a := p.transform.At(0, 0)
	b := p.transform.At(1, 0)
	c := p.transform.At(0, 1)
	d := p.transform.At(1, 1)
	e := p.transform.At(0, 2)
	f := p.transform.At(1, 2)

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

	if p.origin == nil {
		// when no origing yet, wait a bit for first event to fillout
		return nil
	}

	x := p.origin.At(0, 0)
	y := p.origin.At(1, 0)

	k := identity()
	k.Mul(translate(-x, -y, 0), k)
	k.Mul(scale(z), k)
	k.Mul(translate(x, y, 0), k)

	p.transform.Mul(k, p.transform)
	p.setRootTranslation()
	return nil
}

func (p *PanZoomer) handleMouseMove(_ js.Value, args []js.Value) interface{} {
	event := args[0]
	point := getEventPoint(event)

	if p.state != Drag {
		p.origin = point
		return nil
	}

	x := point.At(0, 0)
	y := point.At(1, 0)
	ox := p.origin.At(0, 0)
	oy := p.origin.At(1, 0)

	p.transform.Mul(p.transformBeforeDrag, translate(x-ox, y-oy, 0))
	p.setRootTranslation()
	return nil
}

func (p *PanZoomer) handleMouseDown(_ js.Value, args []js.Value) interface{} {
	p.state = Drag
	event := args[0]
	p.origin = getEventPoint(event)
	p.transformBeforeDrag = mat.DenseCopyOf(p.transform)
	return nil
}

func (p *PanZoomer) handleMouseUp(_ js.Value, _ []js.Value) interface{} {
	p.state = NotActive
	return nil
}

func identity() *mat.Dense {
	return mat.NewDense(3, 3, []float64{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	})
}

func scale(z float64) *mat.Dense {
	return mat.NewDense(3, 3, []float64{
		z, 0, 0,
		0, z, 0,
		0, 0, 1,
	})
}

func translate(x, y, z float64) *mat.Dense {
	return mat.NewDense(3, 3, []float64{
		1, 0, x,
		0, 1, y,
		0, 0, 1,
	})
}
