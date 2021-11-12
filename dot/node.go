package dot

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/nikolaydubina/jsonl-graph/graph"
)

// NodeShape should match Graphviz values
type NodeShape string

const (
	NoneShape   NodeShape = "none"
	RecordShape NodeShape = "record"
)

type Node struct {
	id    string
	shape NodeShape
	label Renderable
}

func (r Node) Render() string {
	s := `"` + r.id + `"` + "\n"
	s += "[\n"
	s += "shape=" + string(r.shape) + "\n"
	s += `label=` + r.label.Render() + "\n" // note, node should wrap to string itself
	s += "]\n"
	return s
}

// BasicNodeLabel is label content for non-colorized Graphviz node
type BasicNodeLabel struct {
	n graph.NodeData
}

func (r BasicNodeLabel) Render() string {
	rows := []string{}
	for k, v := range r.n {
		if k == "id" {
			continue
		}

		if strings.HasSuffix(k, "_url") {
			// URLs tend to be big and clutter dot outputs
			continue
		}

		rows = append(rows, fmt.Sprintf(`{%v\l | %s\r}`, k, Value{v: v}.Render()))
	}

	// this will sort by key, since key is first
	sort.Strings(rows)

	return fmt.Sprintf(`"{ %s | %s }"`, r.n["id"], strings.Join(rows, " | "))
}

// Value coerces to json.Number and tries to avoid adding decimal points to integers
type Value struct {
	v interface{}
}

func (r Value) Render() string {
	if v, ok := r.v.(json.Number); ok {
		if vInt, err := v.Int64(); err == nil {
			return fmt.Sprintf("%d", vInt)
		}
		if v, err := v.Float64(); err == nil {
			return fmt.Sprintf("%.2f", v)
		}
	}
	return fmt.Sprintf("%v", r.v)
}
