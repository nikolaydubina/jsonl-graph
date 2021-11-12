package svg

import (
	"fmt"
	"strings"
)

type Renderable interface {
	Render() string
}

type SVG struct {
	ID          string
	Definitions []Renderable
	Body        Renderable
}

func (s SVG) Render() string {
	defs := make([]string, 0, len(s.Definitions))
	for _, d := range s.Definitions {
		defs = append(defs, d.Render())
	}
	return strings.Join(
		[]string{
			fmt.Sprintf(`<svg id="%s" xmlns="http://www.w3.org/2000/svg" style="width: 100%%; height: 100%%;">`, s.ID),
			`<defs>`,
			strings.Join(defs, "\n"),
			`</defs>`,
			s.Body.Render(),
			`</svg>`,
		},
		"\n",
	)
}

type ArrowDef struct{}

func (s ArrowDef) Render() string {
	return `
		<marker id="arrowhead" markerWidth="10" markerHeight="7" refX="0" refY="3.5" orient="auto">
			<polygon points="0 0, 10 3.5, 0 7" />
		</marker>
	`
}
