package renderer

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
	for fromID, edges := range graphData.Edges {
		if _, ok := graphRender.Edges[fromID]; !ok {
			graphRender.Edges[fromID] = make(map[uint64]*render.Edge, len(edges))
		}

		// check all new data edges
		for toID := range edges {
			// new edge, creating new edge
			if _, ok := graphRender.Edges[fromID][toID]; !ok {
				graphRender.Edges[fromID][toID] = &render.Edge{}
			}
			// existing edge. skipping, no fields to update.
		}
	}

	// delete render graph edges that no longer present
	for idFrom, edges := range graphRender.Edges {
		dEdges, ok := graphData.Edges[idFrom]
		if !ok {
			delete(graphRender.Edges, idFrom)
			continue
		}
		for idTo := range edges {
			if _, ok := dEdges[idTo]; !ok {
				delete(graphRender.Edges[idFrom], idTo)
			}
		}
	}
}
