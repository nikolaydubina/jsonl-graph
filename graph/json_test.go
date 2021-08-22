package graph

import (
	"errors"
	"strings"
	"testing"
)

func TestNodeData(t *testing.T) {
	tests := []struct {
		name     string
		NodeData NodeData
		valid    bool
	}{
		{
			name:     "empty",
			NodeData: NodeData(nil),
			valid:    false,
		},
		{
			name:     "empty non nil map",
			NodeData: NodeData(map[string]interface{}{}),
			valid:    false,
		},
		{
			name:     "no id",
			NodeData: NodeData(map[string]interface{}{"asdf": 123}),
			valid:    false,
		},
		{
			name:     "invalid int id",
			NodeData: NodeData(map[string]interface{}{"id": 123}),
			valid:    false,
		},
		{
			name:     "valid with string id",
			NodeData: NodeData(map[string]interface{}{"id": "asdf"}),
			valid:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.NodeData.IsValid() != tc.valid {
				t.Fail()
			}
		})
	}
}

func TestEdgeData(t *testing.T) {
	tests := []struct {
		name     string
		EdgeData EdgeData
		valid    bool
	}{
		{
			name:     "empty",
			EdgeData: EdgeData(nil),
			valid:    false,
		},
		{
			name:     "empty non nil map",
			EdgeData: EdgeData(map[string]interface{}{}),
			valid:    false,
		},
		{
			name:     "no id",
			EdgeData: EdgeData(map[string]interface{}{"asdf": 123}),
			valid:    false,
		},
		{
			name:     "valid with string id",
			EdgeData: EdgeData(map[string]interface{}{"from": "asdf", "to": "12a3"}),
			valid:    true,
		},
		{
			name:     "invalid int id",
			EdgeData: EdgeData(map[string]interface{}{"from": 123, "to": 456}),
			valid:    false,
		},
		{
			name:     "not valid when to is missing",
			EdgeData: EdgeData(map[string]interface{}{"from": "122"}),
			valid:    false,
		},
		{
			name:     "not valid when from is missing",
			EdgeData: EdgeData(map[string]interface{}{"to": "122"}),
			valid:    false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.EdgeData.IsValid() != tc.valid {
				t.Fail()
			}
		})
	}
}

func TestOr(t *testing.T) {
	e := EdgeData(map[string]interface{}{"from": "123", "to": "456"})
	n := NodeData(map[string]interface{}{"id": "123"})

	tests := []struct {
		name string
		or   orNodeDataEdgeData
		n    *NodeData
		e    *EdgeData
		err  error
	}{
		{
			name: "empty",
			or:   orNodeDataEdgeData(nil),
			err:  errors.New("not EdgeData, not NodeData"),
		},
		{
			name: "empty non nil map",
			or:   orNodeDataEdgeData(map[string]interface{}{}),
			err:  errors.New("not EdgeData, not NodeData"),
		},
		{
			name: "wrong",
			or:   orNodeDataEdgeData(map[string]interface{}{"asdf": 123}),
			err:  errors.New("not EdgeData, not NodeData"),
		},
		{
			name: "EdgeData",
			or:   orNodeDataEdgeData(map[string]interface{}{"from": "123", "to": "456"}),
			e:    &e,
		},
		{
			name: "NodeData",
			or:   orNodeDataEdgeData(map[string]interface{}{"id": "123"}),
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
			err:   "not edge, not node",
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
		{"from": "123", "to": "321"}
		`
		g, err := NewGraphFromJSONLReader(strings.NewReader(input))
		if err != nil {
			t.Errorf("expected no error but got %v", err)
		}

		from := g.IDStorage.Get("123")
		if v, ok := g.Nodes[from]; !ok || v["id"] != "123" {
			t.Errorf("bad NodeData %#v", g.Nodes)
		}

		to := g.IDStorage.Get("321")
		if v, ok := g.Nodes[to]; !ok || v["id"] != "321" {
			t.Errorf("bad NodeData %#v", g.Nodes)
		}

		edge := g.Edges[from][to]
		if edge["from"] != "123" || edge["to"] != "321" {
			t.Errorf("bad EdgeData %#v", edge)
		}
	})
}
