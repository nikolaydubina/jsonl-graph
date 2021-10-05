package app

import (
	"github.com/nikolaydubina/jsonl-graph/graph"
	"github.com/nikolaydubina/jsonl-graph/render"
)

// UpdateRenderGraphWithDataGraph is called when graph data changed
// and we need to update render graph nodes and fields based on new
// data from data graph.
// We have to preserve ids and existing render information.
// For example, preserving positions in Nodes and Paths points in Edges.
func UpdateRenderGraphWithDataGraph(graphData graph.Graph, graphRender render.Graph) {
	// update nodes with new data, preserve rest. add new nodes.
	for id, node := range graphData.Nodes {
		if _, ok := graphRender.Nodes[id]; !ok {
			graphRender.Nodes[id] = &render.Node{}
		}

		graphRender.Nodes[id].NodeData = node
		graphRender.Nodes[id].ID = node.ID()
		graphRender.Nodes[id].Title = node.ID()
	}

	// delete render graph nodes that no longer present
	for id := range graphRender.Nodes {
		if _, ok := graphData.Nodes[id]; !ok {
			delete(graphRender.Nodes, id)
		}
	}

	// update edges with new data, preserve rest. add new edges.
	for e := range graphData.Edges {
		// new edge, creating new edge
		if _, ok := graphRender.Edges[e]; !ok {
			graphRender.Edges[e] = &render.Edge{}
		}
		// existing edge. skipping, no fields to update.
	}

	// delete render graph edges that no longer present
	for e := range graphRender.Edges {
		if _, ok := graphData.Edges[e]; !ok {
			delete(graphRender.Edges, e)
		}
	}
}
