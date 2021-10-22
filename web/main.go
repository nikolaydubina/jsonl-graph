package main

import (
	"github.com/nikolaydubina/jsonl-graph/web/app"
)

func main() {
	app.NewBridge(
		"output-container",
		"svg-jsonl-graph",
		"svg-jsonl-graph-root",
	)

	// do not exit
	c := make(chan bool)
	<-c
}
