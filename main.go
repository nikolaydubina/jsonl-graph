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

// decideOrientation picks orientation. Typically graphs with lots of data in nodes better look in TopDown orientation.
func decideOrientation(g graph.Graph) dot.Orientation {
	for _, n := range g.Nodes {
		if len(n) > 1 {
			return dot.TB
		}
	}
	return dot.LR
}

func main() {
	var colorSchemeFilePath string
	var lrOrientation bool
	var tbOrientation bool

	// TODO: help message with examples
	flag.StringVar(&colorSchemeFilePath, "color-scheme", "", "optional file-path to colorscheme file (e.g. file://basic-colors.json)")
	flag.BoolVar(&lrOrientation, "lr", false, "left-right orientation")
	flag.BoolVar(&tbOrientation, "tb", false, "top-bottom orientation")
	flag.Parse()

	// read graph
	g, err := graph.NewGraphFromJSONL(os.Stdin)
	if err != nil {
		log.Fatalf("can no read graph from json: %s", err)
	}

	// orientation
	if lrOrientation && tbOrientation {
		log.Fatalf("can not have lr and tb orientation at same time")
	}
	var orientation dot.Orientation
	switch {
	case lrOrientation:
		orientation = dot.LR
	case tbOrientation:
		orientation = dot.TB
	default:
		orientation = decideOrientation(g)
	}

	var r renderable

	// default
	r = dot.NewBasicGraph(g, orientation)

	// try get colors
	if colorSchemeFilePath != "" {
		if colorFile, err := getFileFromLocalFiles(colorSchemeFilePath); err == nil {
			var conf color.ColorConfig
			if err := json.Unmarshal(colorFile, &conf); err != nil {
				log.Fatalf("bad color config: %s", err)
			}
			r = dot.NewColoredGraph(g, orientation, conf)
		}
	}

	os.Stdout.WriteString(r.Render())
}
