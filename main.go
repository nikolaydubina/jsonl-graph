package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/nikolaydubina/jsonl-graph/color"
	"github.com/nikolaydubina/jsonl-graph/dot"
	"github.com/nikolaydubina/jsonl-graph/graph"
)

type renderable interface {
	Render() string
}

func getFileFromLocalFiles(path string) ([]byte, error) {
	var t http.Transport
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	c := http.Client{Transport: &t}

	res, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("can not load file at path %s: %w", path, err)
	}
	return io.ReadAll(res.Body)
}

func main() {
	var colorSchemeFilePath string

	// TODO: help message with examples
	flag.StringVar(&colorSchemeFilePath, "color-scheme", "", "optional path to colorscheme file (can be e.g. file://basic-colors.json)")
	flag.Parse()

	g, err := graph.NewGraphFromJSONL(os.Stdin)
	if err != nil {
		log.Fatalf("can no read graph from json: %s", err)
	}

	var r renderable

	// default
	r = dot.NewBasicGraph(g, dot.TB)

	// try get colors
	if colorSchemeFilePath != "" {
		if colorFile, err := getFileFromLocalFiles(colorSchemeFilePath); err == nil {
			var conf color.ColorConfig
			if err := json.Unmarshal(colorFile, &conf); err != nil {
				log.Fatalf("bad color config: %s", err)
			}
			r = dot.NewColoredGraph(g, dot.TB, conf)
		}
	}

	os.Stdout.WriteString(r.Render())
}
