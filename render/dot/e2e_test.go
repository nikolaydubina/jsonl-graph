package dot_test

import (
	// embed
	"bytes"
	_ "embed"
	"io"
	"strings"
	"testing"

	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/render/dot"
)

//go:embed testdata/config.json
var tdConfig []byte

//go:embed testdata/gin.jsonl
var tdGinJSONL string

//go:embed testdata/gin.dot
var tdGinDOT string

//go:embed testdata/gin_basic.dot
var tdGinBasicDOT string

//go:embed testdata/small.jsonl
var tdSmallJSONL string

//go:embed testdata/small.dot
var tdSmallDOT string

type renderer interface {
	Render(params dot.TemplateParams, w io.Writer) error
}

func TestE2E(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		out    string
		config []byte
	}{
		{
			name:   "color",
			in:     tdGinJSONL,
			out:    tdGinDOT,
			config: tdConfig,
		},
		{
			name: "color, basic",
			in:   tdGinJSONL,
			out:  tdGinBasicDOT,
		},
		{
			name: "small",
			in:   tdSmallJSONL,
			out:  tdSmallDOT,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, err := graph.NewGraphFromJSONLReader(strings.NewReader(tc.in))
			if err != nil {
				t.Errorf("%#v", err)
			}

			var r renderer = dot.NewBasicRenderer()
			if tc.config != nil {
				conf, err := dot.NewColorConfig(tc.config)
				if err != nil {
					t.Errorf("%#v", err)
				}
				r = dot.NewColorRenderer(conf)
			}

			var buff bytes.Buffer
			if err := r.Render(dot.TemplateParams{Graph: g}, &buff); err != nil {
				t.Errorf("%#v", err)
			}

			if buff.String() != tc.out {
				t.Errorf("bad output: %v", buff.String())
			}
		})
	}
}
