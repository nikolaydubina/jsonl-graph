package layout

import (
	"fmt"
	"sort"
	"strings"
)

// LayeredGraph is graph with dummy nodes such that there is no long edges.
// Short edge is between nodes in Layers next to each other.
// Long edge is between nodes in 1+ Layers between each other.
// Segment is either a short edge or a long edge.
// Top layer has lowest layer number.
type LayeredGraph struct {
	Segments map[[2]uint64]bool     // segment is an edge in layered graph, can be real edge or piece of fake edge
	Dummy    map[uint64]bool        // fake nodes
	NodeYX   map[uint64][2]int      // node -> {layer, ordering in layer}
	Edges    map[[2]uint64][]uint64 // real long/short edge -> {real, fake, fake, fake, real} nodes
}

func (g LayeredGraph) Layers() [][]uint64 {
	maxY := 0
	for _, yx := range g.NodeYX {
		if yx[0] > maxY {
			maxY = yx[0]
		}
	}

	layers := make([][]uint64, maxY+1)
	for y := 0; y < len(layers); y++ {
		// collect to layer
		for node, yx := range g.NodeYX {
			if yx[0] == y {
				layers[y] = append(layers[y], node)
			}
		}

		// sort within layer
		sort.Slice(layers[y], func(i, j int) bool { return g.NodeYX[layers[y][i]][1] < g.NodeYX[layers[y][j]][1] })
	}

	return layers
}

func (g LayeredGraph) Validate() error {
	for e := range g.Segments {
		from := g.NodeYX[e[0]][0]
		to := g.NodeYX[e[1]][0]
		if from >= to {
			return fmt.Errorf("edge(%v) is wrong direction, got from level(%d) to level(%d)", e, from, to)
		}
	}
	return nil
}

func (g LayeredGraph) String() string {
	out := ""

	out += fmt.Sprintf("fake nodes: %v\n", g.Dummy)

	segments := []string{}
	for e := range g.Segments {
		segments = append(segments, fmt.Sprintf("%d->%d", e[0], e[1]))
	}
	sort.Strings(segments)
	out += fmt.Sprintf("segments: %s\n", strings.Join(segments, " "))

	layers := g.Layers()
	for l, nodes := range layers {
		vs := ""
		for _, node := range nodes {
			vs += fmt.Sprintf(" %d", node)
		}
		out += fmt.Sprintf("%d: %s\n", l, vs)
	}
	return out
}

// NumCrossingsAtLayer between layer and its upper and lower layer.
func (g LayeredGraph) NumCrossingsAtLayer(layer int) int {
	count := 0

	for e1 := range g.Segments {
		for e2 := range g.Segments {
			// both edges from same level, to same next level
			if !(g.NodeYX[e1[0]][0] == g.NodeYX[e2[0]][0] && g.NodeYX[e1[1]][0] == g.NodeYX[e2[1]][0]) {
				continue
			}

			// either top or bottom layer has to be our target layer
			if !(g.NodeYX[e1[0]][0] == layer || g.NodeYX[e1[0]][1] == layer) {
				continue
			}

			// e1   e2
			//    x
			// e2   e1
			if (g.NodeYX[e1[0]][1] < g.NodeYX[e2[0]][1]) && (g.NodeYX[e1[1]][1] > g.NodeYX[e2[1]][1]) {
				count++
				continue
			}

			// e2   e1
			//    x
			// e1   e2
			if (g.NodeYX[e2[0]][1] < g.NodeYX[e1[0]][1]) && (g.NodeYX[e2[1]][1] > g.NodeYX[e1[1]][1]) {
				count++
				continue
			}
		}
	}

	return count
}

func (g LayeredGraph) NumCrossings() int {
	count := 0
	for i := range g.Layers() {
		count += g.NumCrossingsAtLayer(i)
	}
	return count
}

// IsInnerSegment tells when edge is between two Dummy nodes.
func (g LayeredGraph) IsInnerSegment(segment [2]uint64) bool {
	return g.Dummy[segment[0]] && g.Dummy[segment[1]]
}

// UpperNeighbors are nodes in upper layer that are connected to given node.
func (g LayeredGraph) UpperNeighbors(node uint64) []uint64 {
	var nodes []uint64
	for e := range g.Segments {
		if e[1] == node {
			if g.NodeYX[e[0]][0] == (g.NodeYX[e[1]][0] - 1) {
				nodes = append(nodes, e[0])
			}
		}
	}
	return nodes
}

// LowerNeighbors are nodes in lower layer that are connected to given node.
func (g LayeredGraph) LowerNeighbors(node uint64) []uint64 {
	var nodes []uint64
	for e := range g.Segments {
		if e[0] == node {
			if g.NodeYX[e[0]][0] == (g.NodeYX[e[1]][0] - 1) {
				nodes = append(nodes, e[0])
			}
		}
	}
	return nodes
}
