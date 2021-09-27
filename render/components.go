package render

import (
	"encoding/json"
	"fmt"
	"strings"
)

func svg(defs []string, id string, body string) string {
	return strings.Join(
		[]string{
			fmt.Sprintf(`<svg id="%s" xmlns="http://www.w3.org/2000/svg" style="width: 100%%; height: 100%%;">`, id),
			`<defs>`,
			strings.Join(defs, "\n"),
			`</defs>`,
			body,
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

// RenderValue coerces to json.Number and tries to avoid adding decimal points to integers
func RenderValue(v interface{}) string {
	if v, ok := v.(json.Number); ok {
		if vInt, err := v.Int64(); err == nil {
			return fmt.Sprintf("%d", vInt)
		}
		if v, err := v.Float64(); err == nil {
			return fmt.Sprintf("%.2f", v)
		}
	}
	return fmt.Sprintf("%v", v)
}
