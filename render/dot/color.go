package dot

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/nikolaydubina/jsonl-graph/graph"
)

//go:embed templates/color.dot
var colorTemplate string

// ColorScale is sequence of colors and numerical anchors between them
type ColorScale struct {
	Colors []color.RGBA
	Points []float64
}

// ColorConfigVal is configuration for single key on how to color its value
type ColorConfigVal struct {
	ValToColor map[string]color.RGBA `json:"ColorMapping"`
	ColorScale *ColorScale           `json:"ColorScale"`
}

// ColorConfig is config for all keys
type ColorConfig map[string]ColorConfigVal

// NewColorConfigFromFileURL loads from local file like file:///myconfig.json
func NewColorConfigFromFileURL(path string) (ColorConfig, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}

	t := http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	c := http.Client{Transport: &t}

	res, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("can not load colorscheme file at path %s: %w", path, err)
	}
	colorschemeBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("can not read file: %w", err)
	}
	return NewColorConfig(colorschemeBytes)
}

// NewColorConfig loads from string
func NewColorConfig(s []byte) (ColorConfig, error) {
	var conf ColorConfig
	if err := json.Unmarshal(s, &conf); err != nil {
		return nil, fmt.Errorf("can not unmarshal: %w", err)
	}
	return conf, nil
}

// Color returns color for single key based on its value
func (c ColorConfig) Color(k string, v interface{}) color.Color {
	valC, ok := c[k]
	if !ok {
		return color.White
	}

	// first check manual values
	var key string
	if vs, ok := v.(string); ok {
		key = vs
	} else {
		vs, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		key = string(vs)
	}
	if c, ok := valC.ValToColor[key]; ok {
		return c
	}

	// then check scale if present
	if valC.ColorScale != nil && len(valC.ColorScale.Points) > 0 && len(valC.ColorScale.Points) == (len(valC.ColorScale.Colors)-1) {
		if vs, err := strconv.ParseFloat(key, 64); err == nil {
			idx := 0
			for idx < len(valC.ColorScale.Points) && valC.ColorScale.Points[idx] <= vs {
				idx++
			}
			if idx >= len(valC.ColorScale.Colors) {
				idx = len(valC.ColorScale.Colors) - 1
			}
			return valC.ColorScale.Colors[idx]
		}

	}

	return color.White
}

// ColorRenderer contains methods to transform input to Graphviz format
// TODO: consider adding colors in background https://stackoverflow.com/questions/17765301/graphviz-dot-how-to-change-the-colour-of-one-record-in-multi-record-shape
type ColorRenderer struct {
	Template    *template.Template
	ColorConfig ColorConfig
}

// NewColorRenderer initializes template for reuse
func NewColorRenderer(conf ColorConfig) ColorRenderer {

	ret := ColorRenderer{
		ColorConfig: conf,
	}
	ret.Template = template.Must(template.New("colorDotTemplate").Funcs(template.FuncMap{
		"nodeLabelTableColored": ret.RenderLabelTableColored,
	}).Parse(colorTemplate))

	return ret
}

// Render writes graph in Graphviz format to writer
func (c ColorRenderer) Render(params TemplateParams, w io.Writer) error {
	params.UpdateOrientation()
	return c.Template.Execute(w, params)
}

// Color transforms Go color to Graphviz RGBA format
func Color(c color.Color) string {
	r, g, b, a := c.RGBA()
	return fmt.Sprintf("#%x%x%x%x", uint8(r), uint8(g), uint8(b), uint8(a))
}

// RenderLabelTableColored makes graphviz string for a single node with colored table
func (c ColorRenderer) RenderLabelTableColored(n graph.NodeData) string {
	rows := []string{}
	for k, v := range n {
		if k == "id" || strings.HasSuffix(k, "_url") {
			continue
		}

		row := fmt.Sprintf(`
			<tr>
				<td border="1" ALIGN="LEFT">%s</td>
				<td border="1" ALIGN="RIGHT" bgcolor="%s">%s</td>
			</tr>`,
			k,
			Color(c.ColorConfig.Color(k, v)),
			RenderValue(v),
		)

		rows = append(rows, row)
	}

	// this will sort by key, since key is first
	sort.Strings(rows)

	return strings.Join(
		[]string{
			"<<table border=\"0\" cellspacing=\"0\" CELLPADDING=\"6\">",
			fmt.Sprintf(`
				<tr>
					<td port="port0" border="1" colspan="2" ALIGN="CENTER" bgcolor="%s">%s</td>
				</tr>`,
				Color(color.RGBA{R: 200, G: 200, B: 200, A: 200}),
				RenderValue(n["id"]),
			),
			strings.Join(rows, "\n"),
			"</table>>",
		},
		"\n",
	)
}
