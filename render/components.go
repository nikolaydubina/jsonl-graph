package render

import (
	"fmt"
	"strings"
)

func svg(defs []string, body []string) string {
	return strings.Join(
		[]string{
			fmt.Sprintf(`<svg id="graph2" xmlns="http://www.w3.org/2000/svg" style="width: 100%%; height: 100%%;">`),
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
