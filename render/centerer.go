package render

// CenterBox computes coordinates on how to center box.
// Negative box coordinates are allowed.
// This is three step process:
//  1. Move bounding box to start from 0,0.
//  2. Align centers of bounding box and screen.
//  3. Zoom bounding box until it fits one of dimensions of the screen.
//
// You have to fist move by dx and dy, and then apply zoom at center point of screen.
func CenterBox(wscreen, hscreen, minx, miny, maxx, maxy float64) (dx, dy, zoom float64) {
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
	dx += (wscreen - wbox) / 2
	dy += (hscreen - hbox) / 2

	// apply zoom to fit bounding box
	switch {
	case wscreen < hscreen && wbox < hbox:
		zoom = wscreen / wbox
	case wscreen < hscreen && wbox > hbox:
		zoom = hscreen / hbox
	case wscreen > hscreen && wbox < hbox:
		zoom = hscreen / hbox
	case wscreen > hscreen && wbox > hbox:
		zoom = wscreen / wbox
	}

	return dx, dy, zoom
}
