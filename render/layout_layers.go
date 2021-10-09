package render

import (
	"fmt"
	"image"
	"log"

	"github.com/nikolaydubina/jsonl-graph/render/brandeskopf"
	"go.uber.org/multierr"
)

// BasicLayersLayout will assign nodes to levels based by depth of a node.
// Expects directed acyclic graph. TODO: add preprocessing for unordered/cyclic graphs.
// This algorithm requires multiple NP-hard steps. Optimizations would be nice.
// This algorithm attempts to match Kozo Sugiyama algorithm for layers.
type BasicLayersLayout struct {
	MarginY        int
	MarginX        int
	FakeNodeWidth  int
	FakeNodeHeight int
}

func (l BasicLayersLayout) UpdateGraphLayout(g Graph) {
	// assign to layers by depth
	nodeYX := getLayersXY(g)
	if err := validateLayers(g, nodeYX); err != nil {
		log.Printf("got wrong layers: %s", err)
	}
	log.Printf("basic layers: made layers:\n%s", NewLayers(nodeYX))

	// add fake nodes when edges cross layers
	edgeFakeNodes := addFakeNodesLayers(g, nodeYX)
	fakeNodes := getFakeMapFromEdgeFakeNodes(edgeFakeNodes)
	log.Printf("basic layers: added fake nodes(%v) for edges(%v):\n%s", fakeNodes, edgeFakeNodes, NewLayers(nodeYX))

	// find allocation within layers to minimize crosses
	log.Printf("basic layers: number of crossings before optimization: %d", numCrossings(g, nodeYX))
	nodeYX = randomLayersOrderingOptimizer(g, nodeYX, 50)
	log.Printf("basic layers: number of crossings after optimization: %d\n%s", numCrossings(g, nodeYX), NewLayers(nodeYX))

	// horizontal coordinate assignment
	segments := make(map[[2]uint64]bool, len(g.Edges))
	for e := range g.Edges {
		segments[e] = true
	}
	x := brandeskopf.BrandesKopfLayersHorizontalAssignment(
		brandeskopf.LayeredGraph{
			Segments: segments,
			Dummy:    fakeNodes,
			NodeYX:   brandeskopf.NodeYX(nodeYX),
			Layers:   brandeskopf.Layers(NewLayers(nodeYX)),
		},
		l.MarginX,
	)

	// draw real nodes
	yOffset := 0
	for _, nodes := range NewLayers(nodeYX) {
		xOffset := 0
		hMax := 0

		for _, node := range nodes {
			w := 0
			h := 0

			if fakeNodes[node] {
				w = l.FakeNodeWidth
				h = l.FakeNodeHeight
			} else {
				w = g.Nodes[node].Width()
				h = g.Nodes[node].Height()
			}

			g.Nodes[node].LeftBottom.X = x[node]
			g.Nodes[node].LeftBottom.Y = yOffset

			xOffset += w
			if h > hMax {
				hMax = h
			}
		}

		yOffset += hMax + l.MarginY
	}

	// draw real edges and edges through fake nodes
	for e := range g.Edges {
		if ns, hasFake := edgeFakeNodes[e]; hasFake {
			var points []image.Point

			// start from real
			from := *g.Nodes[e[0]]
			points = append(points, from.LeftBottom.Add(image.Point{X: from.Width() / 2, Y: from.Height() / 2}))

			// go through fake
			for _, n := range ns {
				fn := *g.Nodes[n]
				points = append(points, fn.LeftBottom.Add(image.Point{X: fn.Width() / 2, Y: fn.Height() / 2}))
			}

			// end with real
			to := *g.Nodes[e[1]]
			points = append(points, to.LeftBottom.Add(image.Point{X: to.Width() / 2, Y: to.Height() / 2}))

			g.Edges[e] = &Edge{Points: points}
		} else {
			edge := DirectEdge(*g.Nodes[e[0]], *g.Nodes[e[1]])
			g.Edges[e] = &edge
		}
	}

	// delete fake nodes
	for node := range fakeNodes {
		delete(g.Nodes, node)
	}
}

type NodeYX map[uint64][2]int

func getLayersXY(g Graph) NodeYX {
	nodeYX := make(map[uint64][2]int, len(g.Nodes))
	for node := range g.Nodes {
		nodeYX[node] = [2]int{0, 0}
	}

	for _, root := range g.GetRoots() {
		log.Printf("root(%d): %s", root, g.Nodes[root].ID)
		visited := map[uint64]bool{}
		nodeYX[root] = [2]int{0, 0}
		for que := []uint64{root}; len(que) > 0; {
			// pop
			p := que[0]
			if len(que) > 1 {
				que = que[1:]
			} else {
				que = nil
			}

			if visited[p] {
				continue
			}
			visited[p] = true

			// set max depth for each child
			for e := range g.Edges {
				if e[0] == p {
					child := e[1]

					if nodeYX[child][0] < (nodeYX[p][0] + 1) {
						nodeYX[child] = [2]int{nodeYX[p][0] + 1, 0}
					}

					que = append(que, child)
				}
			}
		}
	}

	return nodeYX
}

func validateLayers(g Graph, layers NodeYX) error {
	var errs []error

	for e := range g.Edges {
		from := layers[e[0]][0]
		to := layers[e[1]][0]
		if from >= to {
			errs = append(errs, fmt.Errorf("edge(%v) is wrong direction, got from level(%d) to level(%d)", e, from, to))
		}
	}

	return multierr.Combine(errs...)
}

func addFakeNodesLayers(g Graph, layers NodeYX) (edgeFakeNodes map[[2]uint64][]uint64) {
	edgeFakeNodes = map[[2]uint64][]uint64{}

	for e := range g.Edges {
		from := layers[e[0]][0]
		to := layers[e[1]][0]
		if (to - from) > 1 {
			edgeFakeNodes[e] = []uint64{}
			for i := from + 1; i < to; i++ {
				newID := g.AddNode(&Node{})
				layers[newID] = [2]int{i, 0}
				edgeFakeNodes[e] = append(edgeFakeNodes[e], newID)
			}
		}
	}

	return edgeFakeNodes
}

func getFakeMapFromEdgeFakeNodes(edgeFakeNodes map[[2]uint64][]uint64) map[uint64]bool {
	fakeNodes := map[uint64]bool{}
	for _, nodes := range edgeFakeNodes {
		for _, node := range nodes {
			fakeNodes[node] = true
		}
	}
	return fakeNodes
}

func randomLayersOrderingOptimizer(g Graph, nodeYX NodeYX, tries int) NodeYX {
	var bestNodeYX *NodeYX
	bestnc := -1

	for i := 0; i < tries; i++ {
		currNodeYX := NewLayers(nodeYX).AssingRandomX().ToNodeYX()
		medianLayersOrderingOptimizer(g, currNodeYX)

		currnc := 0

		if bestnc == -1 {
			bestNodeYX = &currNodeYX
			bestnc = numCrossings(g, *bestNodeYX)
		} else {
			currnc = numCrossings(g, currNodeYX)
			if currnc < bestnc {
				bestnc = currnc
				bestNodeYX = &currNodeYX
			}
		}

		log.Printf("random updater nodes levels order: min_crossings(%d) trial(%d) crossings(%d)", bestnc, i, currnc)
	}
	return *bestNodeYX
}

// This heuristic takes medium of upper neighbors.
// Median has property of stable vertical edges.
func medianLayersOrderingOptimizer(g Graph, nodeYX NodeYX) {
	for node, yx := range nodeYX {
		// first level there is no parent. skip
		if yx[0] == 0 {
			continue
		}

		// median of parents
		var parents []uint64
		for e := range g.Edges {
			if e[1] == node {
				parents = append(parents, e[0])
			}
		}

		if len(parents) > 0 {
			median := parents[len(parents)/2]
			nodeYX[node] = [2]int{yx[0], nodeYX[median][1]}
		}
	}

	// reconcile position collisions by rolling increment in layer
	// Layers structure can do this reconciliation, since it does sorting in layers.
	for node, yx := range NewLayers(nodeYX).ToNodeYX() {
		nodeYX[node] = yx
	}
}

func numCrossings(g Graph, nodeYX NodeYX) int {
	count := 0

	for e1 := range g.Edges {
		for e2 := range g.Edges {
			// both edges from same level, to same next level
			if !(nodeYX[e1[0]][0] == nodeYX[e2[0]][0] && nodeYX[e1[1]][0] == nodeYX[e2[1]][0]) {
				continue
			}

			// e1   e2
			//    x
			// e2   e1
			if (nodeYX[e1[0]][1] < nodeYX[e2[0]][1]) && (nodeYX[e1[1]][1] > nodeYX[e2[1]][1]) {
				count++
				continue
			}

			// e2   e1
			//    x
			// e1   e2
			if (nodeYX[e2[0]][1] < nodeYX[e1[0]][1]) && (nodeYX[e2[1]][1] > nodeYX[e1[1]][1]) {
				count++
				continue
			}
		}
	}

	return count
}
