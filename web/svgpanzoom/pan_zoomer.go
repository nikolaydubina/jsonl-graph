package svgpanzoom

import (
	"fmt"
	"log"
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
// Zoom and Dragging is not supported at same time.
// This code is using a lot of JS functions for matrix manipulations.
type PanZoomer struct {
	rootID    string
	zoomScale float64
	state     State
	transform mat.Dense
	origin    mat.Dense
}

func NewPanZoomer(
	rootID string,
	zoomScale float64,
) *PanZoomer {
	return &PanZoomer{
		rootID:    rootID,
		zoomScale: zoomScale,
		transform: *newIdentity(),
	}
}

// parseDOMMatrix will parse JS DOMMatrix spec into Go matrix.
func parseDOMMatrix(val js.Value) *mat.Dense {
	m11 := val.Get("m11").Float()
	m12 := val.Get("m12").Float()
	m13 := val.Get("m13").Float()
	m14 := val.Get("m14").Float()
	m21 := val.Get("m21").Float()
	m22 := val.Get("m22").Float()
	m23 := val.Get("m23").Float()
	m24 := val.Get("m24").Float()
	m31 := val.Get("m31").Float()
	m32 := val.Get("m32").Float()
	m33 := val.Get("m33").Float()
	m34 := val.Get("m34").Float()
	m41 := val.Get("m41").Float()
	m42 := val.Get("m42").Float()
	m43 := val.Get("m43").Float()
	m44 := val.Get("m44").Float()

	return mat.NewDense(4, 4, []float64{
		m11, m12, m13, m14,
		m21, m22, m23, m24,
		m31, m32, m33, m34,
		m41, m42, m43, m44,
	})
}

// TODO: add DOMMouseScroll support
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

func (p *PanZoomer) getRoot() js.Value {
	return js.Global().Get("document").Call("getElementById", p.rootID)
}

func (p *PanZoomer) getEventPoint(event js.Value) *mat.Dense {
	return mat.NewDense(4, 1, []float64{
		event.Get("clientX").Float(),
		event.Get("clientY").Float(),
		0,
		1,
	})
}

// set transformation field to SVG element.
func setCTM(element js.Value, m mat.Dense) {
	a := m.At(0, 0)
	b := m.At(0, 1)
	c := m.At(1, 0)
	d := m.At(1, 1)
	e := m.At(2, 0)
	f := m.At(2, 1)

	s := fmt.Sprintf("matrix(%f,%f,%f,%f,%f,%f)", a, b, c, d, e, f)
	element.Call("setAttribute", "transform", s)
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
	point.Mul(&p.transform, p.getEventPoint(event))

	// Compute new scale matrix in current mouse position
	x := point.At(0, 0)
	y := point.At(1, 0)
	var k mat.Dense
	k = *newIdentity()
	k.Mul(&k, newTranslate(x, y, 0))
	k.Mul(&k, newScale(z))
	k.Mul(&k, newTranslate(-x, -y, 0))

	var ki mat.Dense
	_ = ki.Inverse(&k)

	p.transform.Mul(&p.transform, &ki)

	setCTM(p.getRoot(), p.transform)
	return nil
}

func (p *PanZoomer) handleMouseMove(_ js.Value, args []js.Value) interface{} {
	log.Printf("%#v", p)
	if p.state != Drag {
		return nil
	}

	event := args[0]
	log.Printf("%s", event.String())

	var point mat.Dense
	point.Mul(&p.transform, p.getEventPoint(event))

	x := point.At(0, 0)
	y := point.At(1, 0)
	ox := p.origin.At(0, 0)
	oy := p.origin.At(1, 0)

	var itr mat.Dense
	itr.Inverse(&p.transform)

	translate := newTranslate(x-ox, y-oy, 0)
	p.transform.Mul(&p.transform, translate)

	setCTM(p.getRoot(), p.transform)
	return nil
}

func (p *PanZoomer) handleMouseDown(_ js.Value, args []js.Value) interface{} {
	p.state = Drag
	event := args[0]
	p.origin = *p.getEventPoint(event)
	return nil
}

func (p *PanZoomer) handleMouseUp(_ js.Value, _ []js.Value) interface{} {
	p.state = NotActive
	return nil
}
