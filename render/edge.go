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

// DirectEdge is straight line from center of one node to another.
func DirectEdge(from, to Node) Edge {
	return Edge{
		Points: []image.Point{
			from.LeftBottom.Add(image.Point{X: from.Width() / 2, Y: from.Height() / 2}),
			to.LeftBottom.Add(image.Point{X: to.Width() / 2, Y: to.Height() / 2}),
		},
	}
}
