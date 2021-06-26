package graph

import (
	"errors"
	"strings"
	"testing"
)

func TestNode(t *testing.T) {
	tests := []struct {
		name  string
		node  Node
		valid bool
	}{
		{
			name:  "empty",
			node:  Node(nil),
			valid: false,
		},
		{
			name:  "empty non nil map",
			node:  Node(map[string]interface{}{}),
			valid: false,
		},
		{
			name:  "no id",
			node:  Node(map[string]interface{}{"asdf": 123}),
			valid: false,
		},
		{
			name:  "valid with string id",
			node:  Node(map[string]interface{}{"id": "asdf"}),
			valid: true,
		},
		{
			name:  "valid with int id",
			node:  Node(map[string]interface{}{"id": 123}),
			valid: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.node.IsValid() != tc.valid {
				t.Fail()
			}
		})
	}
}

func TestEdge(t *testing.T) {
	tests := []struct {
		name  string
		edge  Edge
		valid bool
	}{
		{
			name:  "empty",
			edge:  Edge(nil),
			valid: false,
		},
		{
			name:  "empty non nil map",
			edge:  Edge(map[string]interface{}{}),
			valid: false,
		},
		{
			name:  "no id",
			edge:  Edge(map[string]interface{}{"asdf": 123}),
			valid: false,
		},
		{
			name:  "valid with string id",
			edge:  Edge(map[string]interface{}{"from": "asdf", "to": "12a3"}),
			valid: true,
		},
		{
			name:  "valid with int id",
			edge:  Edge(map[string]interface{}{"from": 123, "to": 456}),
			valid: true,
		},
		{
			name:  "not valid when to is missing",
			edge:  Edge(map[string]interface{}{"from": 122}),
			valid: false,
		},
		{
			name:  "not valid when from is missing",
			edge:  Edge(map[string]interface{}{"to": 122}),
			valid: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.edge.IsValid() != tc.valid {
				t.Fail()
			}
		})
	}
}

func TestOr(t *testing.T) {
	e := Edge(map[string]interface{}{"from": 123, "to": 456})
	n := Node(map[string]interface{}{"id": 123})

	tests := []struct {
		name string
		or   orNodeEdge
		n    *Node
		e    *Edge
		err  error
	}{
		{
			name: "empty",
			or:   orNodeEdge(nil),
			err:  errors.New("not edge, not node"),
		},
		{
			name: "empty non nil map",
			or:   orNodeEdge(map[string]interface{}{}),
			err:  errors.New("not edge, not node"),
		},
		{
			name: "wrong",
			or:   orNodeEdge(map[string]interface{}{"asdf": 123}),
			err:  errors.New("not edge, not node"),
		},
		{
			name: "edge",
			or:   orNodeEdge(map[string]interface{}{"from": 123, "to": 456}),
			e:    &e,
		},
		{
			name: "node",
			or:   orNodeEdge(map[string]interface{}{"id": 123}),
			n:    &n,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n, e, err := tc.or.cast()
			if n != nil || tc.n != nil {
				if (*n)["id"] != (*tc.n)["id"] {
					t.Fail()
				}
			}
			if e != nil || tc.e != nil {
				if (*e)["from"] != (*tc.e)["from"] || (*e)["to"] != (*tc.e)["to"] {
					t.Fail()
				}
			}
			if err != nil || tc.err != nil {
				if errors.Is(err, tc.err) {
					t.Errorf("%v but got %v", tc.err, err)
				}
			}
		})
	}
}

func TestParser(t *testing.T) {
	tests := []struct {
		input string
		err   string
	}{
		{
			input: `{"asdf": 123}`,
			err:   "can not cast: not edge, not node",
		},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, err := NewGraphFromJSONLReader(strings.NewReader(tc.input))
			if !strings.Contains(err.Error(), tc.err) {
				t.Errorf("%v but got %v", tc.err, err)
			}
		})
	}

	t.Run("success", func(t *testing.T) {
		input := `
		{"id": "123"}
		{"from": "123", "to": "231"}
		`
		g, err := NewGraphFromJSONLReader(strings.NewReader(input))
		if err != nil {
			t.Errorf("expected no error but got %v", err)
		}
		if g.Nodes[0]["id"] != "123" {
			t.Errorf("bad node %#v", g.Nodes[0])
		}
		if g.Edges[0]["from"] != "123" || g.Edges[0]["to"] != "231" {
			t.Errorf("bad edge %#v", g.Edges[0])
		}
	})
}
