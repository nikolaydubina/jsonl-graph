package dot

import (
	// embed
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/nikolaydubina/jsonl-graph/graph"
)

//go:embed templates/basic.dot
var basicTemplate string

// BasicRenderer contains methods to transform input to  format
// TODO: consider adding colors in background https://stackoverflow.com/questions/17765301/-dot-how-to-change-the-colour-of-one-record-in-multi-record-shape
type BasicRenderer struct {
	Template *template.Template
}

// NewBasicRenderer initializes template for reuse
func NewBasicRenderer() BasicRenderer {
	return BasicRenderer{
		Template: template.Must(template.New("basicDotTemplate").Funcs(template.FuncMap{
			"nodeLabelBasic": RenderBasicLabel,
		}).Parse(basicTemplate)),
	}
}

// Render writes graph in  format to writer
func (g BasicRenderer) Render(params TemplateParams, w io.Writer) error {
	params.UpdateOrientation()
	return g.Template.Execute(w, params)
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

// RenderBasicLabel makes  string for a single node
// This is pretty complex to write in Go template language due to map structure.
func RenderBasicLabel(n graph.Node) string {
	rows := []string{}
	for k, v := range n {
		if k == "id" {
			continue
		}

		if strings.HasSuffix(k, "_url") {
			// URLs tend to be big and clutter dot outputs
			continue
		}

		rows = append(rows, fmt.Sprintf(`{%v\l | %s\r}`, k, RenderValue(v)))
	}

	// this will sort by key, since key is first
	sort.Strings(rows)

	return fmt.Sprintf("{ %s | %s }", n["id"], strings.Join(rows, " | "))
}
