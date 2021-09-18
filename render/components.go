package render

import (
	"fmt"
	"image"
	"strings"
)

func svg(defs []string, body []string, leftBottom, rightTop image.Point) string {
	return strings.Join(
		[]string{
			fmt.Sprintf(`<svg id="graph" xmlns="http://www.w3.org/2000/svg" viewBox="%d %d %d %d" style="width: 100%%; height: 100%%;">`, leftBottom.X, leftBottom.Y, rightTop.X, rightTop.Y),
			`<defs>`,
			strings.Join(defs, "\n"),
			`</defs>`,
			strings.Join(body, "\n"),
			`</svg>`,
		},
		"\n",
	)
}

func arrowDef() string {
	return `
		<marker id="arrowhead" markerWidth="10" markerHeight="7" refX="0" refY="3.5" orient="auto">
			<polygon points="0 0, 10 3.5, 0 7" />
		</marker>
	`
}
