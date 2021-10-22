package layout

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"go.uber.org/multierr"
)

// Does not store exact XY coordinates.
// Does not store paths for edges.
type LayeredGraph struct {
	Segments map[[2]uint64]bool // segment is an edge in layered graph, can be real edge or piece of fake edge
	Dummy    map[uint64]bool    // fake nodes
	NodeYX   map[uint64][2]int  // node -> {layer, ordering in layer}
}

// Expects that graph g does not have cycles.
func NewLayeredGraph(g Graph) LayeredGraph {
	nodeYX := make(map[uint64][2]int, len(g.Nodes))

	for _, root := range roots(g) {
		nodeYX[root] = [2]int{0, 0}
		for que := []uint64{root}; len(que) > 0; {
			// pop
			p := que[0]
			if len(que) > 1 {
				que = que[1:]
			} else {
				que = nil
			}

			// set max depth for each child
			for e := range g.Edges {
				if parent, child := e[0], e[1]; parent == p {
					if l := nodeYX[parent][0] + 1; l > nodeYX[child][0] {
						nodeYX[child] = [2]int{l, 0}
					}
					que = append(que, child)
				}
			}
		}
	}

	// segments
	segments := map[[2]uint64]bool{}
	for e := range g.Edges {
		segments[e] = true
	}

	lg := LayeredGraph{
		NodeYX:   nodeYX,
		Segments: segments,
		Dummy:    map[uint64]bool{},
	}
	log.Printf("layered graph: %s\n", lg)

	if err := lg.Validate(); err != nil {
		panic(fmt.Sprintf("got wrong layers: %s", err))
	}

	return lg
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
	var errs []error

	for e := range g.Segments {
		from := g.NodeYX[e[0]][0]
		to := g.NodeYX[e[1]][0]
		if from >= to {
			errs = append(errs, fmt.Errorf("edge(%v) is wrong direction, got from level(%d) to level(%d)", e, from, to))
		}
	}

	return multierr.Combine(errs...)
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

// AddFakeNodes and edges to the graph and add them to layers structure.
func (g LayeredGraph) AddFakeNodes() {
	var maxNodeID uint64
	for e := range g.Segments {
		if e[0] > maxNodeID {
			maxNodeID = e[0]
		}
		if e[1] > maxNodeID {
			maxNodeID = e[1]
		}
	}
	nextFakeNodeID := maxNodeID + 1

	g.Dummy = map[uint64]bool{}

	for e := range g.Segments {
		from := g.NodeYX[e[0]][0]
		to := g.NodeYX[e[1]][0]

		if (to - from) > 1 {
			// edge to first fake node
			g.Segments[[2]uint64{e[0], nextFakeNodeID}] = true

			for layer := from + 1; layer < to; layer++ {
				g.Dummy[nextFakeNodeID] = true

				g.NodeYX[nextFakeNodeID] = [2]int{layer, 0}

				// edge between fakes
				if (layer > (from + 1)) && (layer < (to - 1)) {
					g.Segments[[2]uint64{nextFakeNodeID - 1, nextFakeNodeID + 1}] = true
				}

				nextFakeNodeID++
			}

			// edge to last fake node
			g.Segments[[2]uint64{nextFakeNodeID - 1, e[1]}] = true
		}
	}
}
