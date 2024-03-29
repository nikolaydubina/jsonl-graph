package graph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/nikolaydubina/multiline-jsonl/mjsonl"
)

// NodeData can have any fields, only `id` is reserved.
// This is JSON representation of node.
type NodeData map[string]interface{}

func (n NodeData) IsValid() bool {
	if len(n) == 0 {
		return false
	}
	return n.ID() != ""
}

func (n NodeData) ID() string {
	v, _ := n["id"].(string)
	return v
}

// EdgeData can have any fields, only `from`, `to`, `label` is reserved.
// This is JSON representation of edge.
type EdgeData map[string]interface{}

func (e EdgeData) IsValid() bool {
	if len(e) == 0 {
		return false
	}
	return e.From() != "" && e.To() != ""
}

func (e EdgeData) From() string {
	v, _ := e["from"].(string)
	return v
}

func (e EdgeData) To() string {
	v, _ := e["to"].(string)
	return v
}

func (e EdgeData) Label() string {
	v, _ := e["label"].(string)
	return v
}

// union type, can be either
type orNodeDataEdgeData map[string]interface{}

func (c orNodeDataEdgeData) cast() (*NodeData, *EdgeData, error) {
	if n := NodeData(c); n.IsValid() {
		return &n, nil, nil
	}
	if e := EdgeData(c); e.IsValid() {
		return nil, &e, nil
	}
	return nil, nil, errors.New("not edge, not node")
}

// Graph is graph structure
type Graph struct {
	Nodes     map[uint64]NodeData
	Edges     map[[2]uint64]EdgeData
	IDStorage MapIDStorage
}

// NewGraph initializes internal structures for empty graph.
func NewGraph() Graph {
	return Graph{
		Nodes:     map[uint64]NodeData{},
		Edges:     map[[2]uint64]EdgeData{},
		IDStorage: NewMapIDStorage(),
	}
}

// AddNode to graph and overwrite if exists.
func (g Graph) AddNode(node NodeData) {
	id := g.IDStorage.Get(node.ID())
	if id == 0 {
		id = g.IDStorage.Add(node.ID())
	}
	g.Nodes[id] = node
}

// AddEdge to graph and overwrite if exists.
// Generate nodes if not present.
func (g Graph) AddEdge(edge EdgeData) {
	fromID := g.IDStorage.Get(edge.From())
	if fromID == 0 {
		g.AddNode(NodeData{"id": edge.From()})
		fromID = g.IDStorage.Get(edge.From())
	}

	toID := g.IDStorage.Get(edge.To())
	if toID == 0 {
		g.AddNode(NodeData{"id": edge.To()})
		toID = g.IDStorage.Get(edge.To())
	}

	g.Edges[[2]uint64{fromID, toID}] = edge
}

// NewGraphFromJSONL parses multiple JSON objects separated by new line.
// JSON objects do not have to be single line.
// Objects are separated by new lines and whitespaces
func NewGraphFromJSONL(r io.Reader) (Graph, error) {
	g := NewGraph()

	scanner := bufio.NewScanner(r)
	scanner.Split(mjsonl.SplitMultilineJSONL)

	for scanner.Scan() {
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		decoder.UseNumber()

		var nodeEdge orNodeDataEdgeData
		if err := decoder.Decode(&nodeEdge); err != nil {
			continue
		}

		node, edge, err := nodeEdge.cast()
		if err != nil {
			return g, fmt.Errorf("can not cast: %w", err)
		}

		switch {
		case node != nil:
			g.AddNode(*node)
		case edge != nil:
			g.AddEdge(*edge)
		}
	}

	return g, scanner.Err()
}

func (g Graph) String() string {
	out := fmt.Sprintf("nodes(%d) edges(%d)\n", len(g.Nodes), len(g.Edges))

	nodes := []string{}
	for id, node := range g.Nodes {
		nodes = append(nodes, fmt.Sprintf("node(%d): %s", id, node.ID()))
	}
	sort.Strings(nodes)
	out += strings.Join(nodes, "\n") + "\n"

	edges := []string{}
	for e := range g.Edges {
		edges = append(edges, fmt.Sprintf("edge(%d -> %d): %s -> %s", e[0], e[1], g.Nodes[e[0]].ID(), g.Nodes[e[1]].ID()))
	}
	sort.Strings(edges)
	out += strings.Join(edges, "\n") + "\n"

	return out
}
