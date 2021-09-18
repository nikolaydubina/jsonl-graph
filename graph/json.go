package graph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// EdgeData can have any fields, only `from` and `to` is reserved.
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
	Edges     map[uint64]map[uint64]EdgeData
	IDStorage MapIDStorage
}

// NewGraph initializes internal structures for empty graph.
func NewGraph() Graph {
	return Graph{
		Nodes:     map[uint64]NodeData{},
		Edges:     map[uint64]map[uint64]EdgeData{},
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

	// generate maps if not present
	if _, ok := g.Edges[fromID]; !ok {
		g.Edges[fromID] = map[uint64]EdgeData{}
	}

	g.Edges[fromID][toID] = edge
}

// ReplaceFrom will move data from other graph while preserving
// IDs from nodes that match "id", "from", "to" keys.
// Nodes and Edges not found in other graph will be removed.
func (g Graph) ReplaceFrom(other Graph) {
	for _, node := range other.Nodes {
		g.AddNode(node)
	}
	for _, edges := range other.Edges {
		for _, edge := range edges {
			g.AddEdge(edge)
		}
	}

	// delete nodes not in other
	for id, node := range g.Nodes {
		// not found
		if other.IDStorage.Get(node.ID()) == 0 {
			delete(g.Nodes, id)
		}
	}

	// delete edges not in other
	for fromID, edges := range g.Edges {
		otherFromID := other.IDStorage.Get(g.Nodes[fromID].ID())
		if otherFromID == 0 || len(other.Edges[otherFromID]) == 0 {
			delete(g.Edges, fromID)
			continue
		}

		for toID := range edges {
			otherToID := other.IDStorage.Get(g.Nodes[toID].ID())
			if _, ok := other.Edges[otherFromID][otherToID]; !ok {
				delete(g.Edges[fromID], toID)
			}
		}
	}
}

// NewGraphFromJSONLReader parses JSONL from reader into a graph.
// TODO: read multiline JSON objects.
func NewGraphFromJSONLReader(r io.Reader) (Graph, error) {
	scanner := bufio.NewScanner(r)

	g := NewGraph()

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
