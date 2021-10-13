package render

import (
	"fmt"
	"image"
	"strings"
)

// Edge is polylines of straight lines going through all points.
type Edge struct {
	Points []image.Point
}

func (e Edge) Render() string {
	var points []string
	for _, point := range e.Points {
		points = append(points, fmt.Sprintf("%d,%d", point.X, point.Y))
	}
	return fmt.Sprintf(`<polyline style="fill:none;stroke-width:1;stroke:black;" points="%s"></polyline>`, strings.Join(points, " "))
}
