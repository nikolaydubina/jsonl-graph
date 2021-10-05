package render

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"sort"

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
	// 1. assign to layers by depth
	nodeYX := getLayersXY(g)
	if err := validateLayers(g, nodeYX); err != nil {
		log.Printf("got wrong layers: %s", err)
	}
	log.Printf("basic layers: made layers:\n%s", NewLayers(nodeYX))

	// 2. add fake nodes when edges cross layers
	edgeFakeNodes := addFakeNodesLayers(g, nodeYX)
	fakeNodes := getFakeMapFromEdgeFakeNodes(edgeFakeNodes)
	log.Printf("basic layers: added fake nodes(%v) for edges(%v):\n%s", fakeNodes, edgeFakeNodes, NewLayers(nodeYX))

	// 3. find allocation within layers to minimize crosses
	log.Printf("basic layers: number of crossings before optimization: %d", numCrossings(g, nodeYX))
	nodeYX = NewLayers(nodeYX).AssingRandomXLayers().ToNodeYX()
	avgUpdateLayersX(g, nodeYX)
	log.Printf("basic layers: number of crossings after optimization: %d\n%s", numCrossings(g, nodeYX), NewLayers(nodeYX))

	// 4. draw real nodes
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

			g.Nodes[node].LeftBottom.X = xOffset
			g.Nodes[node].LeftBottom.Y = yOffset

			xOffset += w + l.MarginX
			if h > hMax {
				hMax = h
			}
		}

		yOffset += hMax + l.MarginY
	}

	// 5. minimize edges betwen layers length
	// TODO

	// 6. draw real edges and edges through fake nodes
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

	// 7. delete fake nodes
	for node := range fakeNodes {
		delete(g.Nodes, node)
	}
}

type NodeYX map[uint64][2]int

func getLayersXY(g Graph) NodeYX {
	layers := make(map[uint64][2]int, len(g.Nodes))

	for _, root := range g.GetRoots() {
		//log.Printf("get layers: root(%d) %s", root, g.Nodes[root].Title)

		que := []uint64{root}
		visited := make(map[uint64]bool, len(g.Nodes))
		for len(que) > 0 {
			// pop
			p := que[0]
			if len(que) > 1 {
				que = que[1:]
			} else {
				que = nil
			}
			//log.Printf("get layers: bfs: p(%d) que(%v) visited(%v) layers(%v)", p, que, visited, layers)

			if visited[p] {
				continue
			}
			visited[p] = true

			// if first time, set level to 0
			if _, ok := layers[p]; !ok {
				layers[p] = [2]int{0, 0}
			}

			// update current node to max of its parents
			hParent := -1
			for e := range g.Edges {
				// edge to current node, update layers depth
				if e[1] == p {
					if v := layers[e[0]]; v[0] >= hParent {
						hParent = v[0]
					}
				}

				// edge from current node, add children to que
				if e[0] == p && !visited[e[1]] {
					que = append(que, e[1])
				}
			}
			layers[p] = [2]int{hParent + 1, 0}
		}
	}

	return layers
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

// This heuristic will take avg of parents x per each node.
// Will assign random to roots.
func avgUpdateLayersX(g Graph, nodeYX NodeYX) {
	for node, yx := range nodeYX {
		// first level there is no parent. skip
		if yx[0] == 0 {
			continue
		}

		// avg location of parents
		totalParentX := 0
		numParents := 0
		for e := range g.Edges {
			// parent
			if e[1] == node {
				totalParentX += nodeYX[e[0]][1]
				numParents++
			}
		}

		if numParents > 0 {
			nodeYX[node] = [2]int{yx[0], totalParentX / numParents}
		}
	}

	// reconcile position collisions by rolling increment in layer
	// Layers structure can do this reconciliation, since it does sorting in layers.
	for node, yx := range NewLayers(nodeYX).ToNodeYX() {
		nodeYX[node] = yx
	}
}

// Layers is more easy form of working with layers and x coordinate of nodes.
// Useful for printing.
// Example:
// 0: 1 8 11
// 1: 5 2
// 2: 11 2 3
type Layers [][]uint64

func NewLayers(nodeYX NodeYX) Layers {
	numLayers := 0
	for _, yx := range nodeYX {
		if yx[0] > numLayers {
			numLayers = yx[0]
		}
	}

	layers := make([][]uint64, numLayers+1)
	for y := 0; y < len(layers); y++ {
		// collect to layer
		for node, yx := range nodeYX {
			if yx[0] == y {
				layers[y] = append(layers[y], node)
			}
		}

		// sort within layer
		sort.Slice(layers[y], func(i, j int) bool { return nodeYX[layers[y][i]][1] < nodeYX[layers[y][j]][1] })
	}

	return layers
}

func (l Layers) ToNodeYX() NodeYX {
	nodeYX := map[uint64][2]int{}
	for y, layer := range l {
		for x, node := range layer {
			nodeYX[node] = [2]int{y, x}
		}
	}
	return nodeYX
}

func (l Layers) AssingRandomXLayers() Layers {
	for i := range l {
		n := len(l[i])
		ordered := make([]uint64, n)
		for from, to := range rand.Perm(n) {
			ordered[to] = uint64(l[i][from])
		}
		copy(l[i], ordered)
	}
	return l
}

func (s Layers) String() string {
	out := ""
	for l, nodes := range s {
		vs := ""
		for _, node := range nodes {
			vs += fmt.Sprintf(" %d", node)
		}
		out += fmt.Sprintf("%d: %s\n", l, vs)
	}
	return out
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
