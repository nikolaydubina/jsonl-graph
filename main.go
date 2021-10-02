// This script transforms JSONL graph from stdin to dot in stdout
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/render/dot"
)

type renderer interface {
	Render(params dot.TemplateParams, w io.Writer) error
}

func main() {
	var colorSchemeFilePath string

	flag.StringVar(&colorSchemeFilePath, "color-scheme", "", "optional path to colorscheme file (can be e.g. file://basic-colors.json)")
	flag.Parse()

	g, err := graph.NewGraphFromJSONL(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var r renderer = dot.NewBasicRenderer()
	if colorSchemeFilePath != "" {
		conf, err := dot.NewColorConfigFromFileURL(colorSchemeFilePath)
		if err != nil {
			log.Fatalf("can not load config with error, will fallback: %s\n", err)
		}
		r = dot.NewColorRenderer(conf)
	}

	if err := r.Render(dot.TemplateParams{Graph: g}, os.Stdout); err != nil {
		log.Printf("fallback renderer got error: %s\n", err)
	}
}
