package layout

import (
	"log"
	"math/rand"
	"sort"
)

type LayerOrderingInitializer interface {
	Init(segments map[[2]uint64]bool, layers [][]uint64)
}

type LayerOrderingOptimizer interface {
	Optimize(segments map[[2]uint64]bool, layers [][]uint64, idx int, downUp bool)
}

// WarfieldOrderingOptimizer is heuristic based strategy for ordering optimization.
// Goes up and down number of iterations across all layers.
// Considers upper and lower fixed and permutes ordering in layer.
// Used in Graphviz/dot.
type WarfieldOrderingOptimizer struct {
	Epochs                   int
	LayerOrderingInitializer LayerOrderingInitializer
	LayerOrderingOptimizer   LayerOrderingOptimizer
}

func (o WarfieldOrderingOptimizer) Optimize(g Graph, lg LayeredGraph) {
	// layers is temporary layers
	layers := lg.Layers()
	o.LayerOrderingInitializer.Init(lg.Segments, layers)

	bestN := -1
	bestLayers := newLayersFrom(layers)

	for t := 0; t < o.Epochs; t++ {
		downUp := (t % 2) == 0
		for i := range layers {
			j := i
			if downUp {
				j = len(layers) - 1 - i
			}
			o.LayerOrderingOptimizer.Optimize(lg.Segments, layers, j, downUp)
		}

		N := numCrossings(lg.Segments, layers)
		if bestN < 0 || N < bestN {
			bestN = N
			copyLayers(bestLayers, layers)
		}
		log.Printf("warfield ordering optimizer:\t epoch(%d)\t best(%d)\t current(%d)\n", t, bestN, N)
	}

	// store to graph
	for y, layer := range bestLayers {
		for x, node := range layer {
			lg.NodeYX[node] = [2]int{y, x}
		}
	}
}

// BFSOrderingInitializer will set order in each layer by traversing BFS from roots.
type BFSOrderingInitializer struct{}

func (o BFSOrderingInitializer) Init(segments map[[2]uint64]bool, layers [][]uint64) {
	// accumulate where node can lead to
	fromNodeToNodes := map[uint64][]uint64{}
	for e := range segments {
		if _, ok := fromNodeToNodes[e[0]]; !ok {
			fromNodeToNodes[e[0]] = []uint64{}
		}
		fromNodeToNodes[e[0]] = append(fromNodeToNodes[e[0]], e[1])
	}

	// get roots
	hasParent := map[uint64]bool{}
	for e := range segments {
		hasParent[e[1]] = true
	}
	var roots []uint64
	for e := range segments {
		if _, ok := hasParent[e[1]]; !ok {
			roots = append(roots, e[1])
		}
	}

	// BFS starting with roots, keeping order when node is visited
	cnt := 0
	que := roots
	ord := map[uint64]int{}
	for len(que) > 0 {
		p := que[0]
		if len(que) > 1 {
			que = que[1:]
		} else {
			que = nil
		}

		if _, ok := ord[p]; ok {
			continue
		}

		ord[p] = cnt
		cnt++

		que = append(que, fromNodeToNodes[p]...)
	}

	for l := range layers {
		sort.Slice(layers[l], func(i, j int) bool { return ord[layers[l][i]] < ord[layers[l][j]] })
	}
}

// RandomLayerOrderingInitializer assigns random ordering in each layer
type RandomLayerOrderingInitializer struct{}

func (o RandomLayerOrderingInitializer) Init(_ map[[2]uint64]bool, layers [][]uint64) {
	for i := range layers {
		l := layers[i]
		rand.Shuffle(len(l), func(a, b int) { l[a], l[b] = l[b], l[a] })
	}
}

// WMedianOrderingOptimizer takes median of upper (or lower) level neighbors for each node in layer.
// Median has property of stable vertical edges which is especially useful for "long" edges (fake nodes).
// Eades and Wormald, 1994
// This is used in dot/Graphviz, Figure 3-2 in Graphviz dot paper TSE93.
type WMedianOrderingOptimizer struct{}

func (o WMedianOrderingOptimizer) Optimize(segments map[[2]uint64]bool, layers [][]uint64, idx int, downUp bool) {
	w := map[uint64]float64{}
	for i, x := range layers[idx] {
		var nidxs []int
		if downUp {
			nidxs = lowerNeighbors(segments, layers, idx, i)
		} else {
			nidxs = upperNeighbors(segments, layers, idx, i)
		}

		P := make([]float64, len(nidxs))
		for i, v := range nidxs {
			P[i] = float64(v)
		}

		w[x] = median(P)
	}
	sort.Slice(layers[idx], func(i, j int) bool { return w[layers[idx][i]] < w[layers[idx][j]] })
}

func lowerNeighbors(segments map[[2]uint64]bool, layers [][]uint64, y int, x int) []int {
	if y == (len(layers) - 1) {
		return nil
	}
	var idxs []int
	for ...
	return idxs
}

func upperNeighbors(segments map[[2]uint64]bool, layers [][]uint64, y int, x int) []int {
	if y == 0 {
		return nil
	}
	var idxs []int
	for ...
	return idxs
}

func median(P []float64) float64 {
	m := len(P) / 2
	switch {
	case len(P) == 0:
		return -1
	case len(P)%2 == 1:
		return P[m]
	case len(P) == 2:
		return (P[0] + P[1]) / 2
	default:
		left := P[m-1] - P[0]
		right := P[len(P)-1] - P[m]
		return (P[m-1]*right + P[m]*left) / (left + right)
	}
}

// RandomLayerOrderingOptimizer picks best out of epochs random orderings.
type RandomLayerOrderingOptimizer struct {
	Epochs int
}

func (o RandomLayerOrderingOptimizer) Optimize(segments map[[2]uint64]bool, layers [][]uint64, idx int, _ bool) {
	bestN := -1
	layer := make([]uint64, len(layers[idx]))
	copy(layer, layers[idx])

	for i := 0; i < o.Epochs; i++ {
		rand.Shuffle(len(layer), func(a, b int) { layer[a], layer[b] = layer[b], layer[a] })

		N := 0
		if idx > 0 {
			N += numCrossingsBetweenLayers(segments, layers[idx-1], layers[idx])
		}
		if (idx + 1) < len(layers) {
			N += numCrossingsBetweenLayers(segments, layers[idx], layers[idx+1])
		}

		if bestN < 0 || N < bestN {
			bestN = N
			copy(layers[idx], layer)
		}
	}
}

// time: O(ntop ^ 2 * nbot ^ 2)
// memory: O(1)
func numCrossingsBetweenLayers(segments map[[2]uint64]bool, ltop, lbottom []uint64) int {
	count := 0

	// e1 top is always to the left from e2 top
	for idxE1T, e1t := range ltop {
		for idxE1B, e1b := range lbottom {
			if _, ok := segments[[2]uint64{e1t, e1b}]; !ok {
				continue
			}
			for idxE2T := idxE1T + 1; idxE2T < len(ltop); idxE2T++ {
				for idxE2B := 0; idxE2B < len(lbottom); idxE2B++ {
					if _, ok := segments[[2]uint64{ltop[idxE2T], lbottom[idxE2B]}]; !ok {
						continue
					}

					// e1   e2
					//    x
					// e2   e1
					if idxE1B > idxE2B {
						count++
					}
				}
			}
		}
	}

	return count
}

// time: O(?)
// memory: O(1)
func numCrossings(segments map[[2]uint64]bool, layers [][]uint64) int {
	count := 0
	for i := range layers {
		if i == 0 {
			continue
		}
		count += numCrossingsBetweenLayers(segments, layers[i-1], layers[i])
	}
	return count
}
