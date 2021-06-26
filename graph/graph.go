package graph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Node can have any fields, only `id` is reserved.
type Node map[string]interface{}

// IsValid checks that node is valid
func (n Node) IsValid() bool {
	if len(n) == 0 {
		return false
	}
	if _, ok := n["id"]; !ok {
		return false
	}
	// TODO: check that value in map is scalar
	return true
}

// Edge can have any fields, only `from` and `to` is reserved.
type Edge map[string]interface{}

// IsValid checks that edge is valid
func (e Edge) IsValid() bool {
	if len(e) == 0 {
		return false
	}
	if _, ok := e["from"]; !ok {
		return false
	}
	if _, ok := e["to"]; !ok {
		return false
	}
	// TODO: check that value in map is scalar
	return true
}

// union type, can be either
type orNodeEdge map[string]interface{}

func (c orNodeEdge) cast() (*Node, *Edge, error) {
	if n := Node(c); n.IsValid() {
		return &n, nil, nil
	}
	if e := Edge(c); e.IsValid() {
		return nil, &e, nil
	}
	return nil, nil, errors.New("not edge, not node")
}

// Graph is graph structure
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// NewGraphFromJSONLReader parses JSONL from reader into a  graph
func NewGraphFromJSONLReader(r io.Reader) (Graph, error) {
	scanner := bufio.NewScanner(r)

	g := Graph{}
	for scanner.Scan() {
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		decoder.UseNumber()

		var nodeEdge orNodeEdge
		if err := decoder.Decode(&nodeEdge); err != nil {
			return g, fmt.Errorf("can not decode to either node or edge: %w", err)
		}

		node, edge, err := nodeEdge.cast()
		if err != nil {
			return g, fmt.Errorf("can not cast: %w", err)
		}

		switch {
		case node != nil:
			g.Nodes = append(g.Nodes, *node)
		case edge != nil:
			g.Edges = append(g.Edges, *edge)
		default:
			return g, errors.New("both edge and node are nil")
		}
	}

	return g, scanner.Err()
}
