package dot

import "github.com/nikolaydubina/jsonl-graph/graph"

// TemplateParams is data passed to template
type TemplateParams struct {
	Graph       graph.Graph
	Orientation string
}

// UpdateOrientation sets orientation depending on node size.
// If set, then will not update.
// This is to make graph look better.
func (c *TemplateParams) UpdateOrientation() {
	if c.Orientation != "" {
		return
	}

	c.Orientation = "TB"
	hasDetails := false
	for _, n := range c.Graph.Nodes {
		if len(n) > 1 {
			hasDetails = true
		}
	}
	if !hasDetails {
		c.Orientation = "LR"
	}
}
