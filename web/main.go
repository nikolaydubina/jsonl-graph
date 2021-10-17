package main

import (
	"github.com/nikolaydubina/jsonl-graph/web/app"
)

func main() {
	c := make(chan bool)

	app.NewBridge(
		"output-container",
		"svg-jsonl-graph",
		"svg-jsonl-graph-root",
	)

	<-c
}
