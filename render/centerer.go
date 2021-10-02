package render

// CenterBox computes coordinates on how to center box.
// Negative box coordinates are allowed.
//
// This is three step process:
//  1. Move bounding box to start from 0,0.
//  2. Align centers of bounding box and screen.
//  3. Zoom boundign box until it fits one of dimensions of the screen.
func CenterBox(wScreen, hScreen, minx, miny, maxx, maxy float64) (dx, dy, zoom float64) {
	dx = 0.0
	dy = 0.0
	zoom = 1.0

	// move to 0,0 beggining of graph
	if minx < 0 {
		dx += -float64(minx)
	}
	if miny < 0 {
		dy += -float64(miny)
	}

	// graph bounding box dimensions
	wbox := maxx - minx
	hbox := maxy - miny

	// align centers of graph bounding box and screen
	dx += (wScreen - wbox) / 2
	dy += (hScreen - hbox) / 2

	// apply zoom to fit graph
	switch {
	case wScreen < hScreen:
		zoom = hScreen / hbox
	case wScreen > hScreen:
		zoom = wScreen / wbox
	}

	return dx, dy, zoom
}
