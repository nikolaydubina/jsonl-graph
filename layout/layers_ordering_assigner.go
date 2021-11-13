package layout

import (
	"math/rand"
)

type LayerOrderingOptimizer interface {
	Optimize(segments map[[2]uint64]bool, layers [][]uint64, idx int)
}

// LBLOrderingOptimizer is layer-by-layer sweep ordering optimizer.
// Goes up and down number of iterations across all layers.
// Considers upper and lower fixed and permutes ordering in layer.
type LBLOrderingOptimizer struct {
	Epochs                 int
	LayerOrderingOptimizer LayerOrderingOptimizer
}

func (o LBLOrderingOptimizer) Optimize(g Graph, lg LayeredGraph) {
	layers := lg.Layers()

	for i := range layers {
		randomizeLayer(layers[i])
	}

	for t := 0; t < o.Epochs; t++ {
		for i := range layers {
			j := i
			if (t % 2) == 0 {
				j = len(layers) - 1 - i
			}
			o.LayerOrderingOptimizer.Optimize(lg.Segments, layers, j)
		}
	}

	for y, layer := range layers {
		for x, node := range layer {
			lg.NodeYX[node] = [2]int{y, x}
		}
	}
}

// RandomLayerOrderingOptimizer picks best out of epochs random orderings.
// Will store inplace in destination best result.
type RandomLayerOrderingOptimizer struct {
	Epochs int
}

func (o RandomLayerOrderingOptimizer) Optimize(segments map[[2]uint64]bool, layers [][]uint64, idx int) {
	if idx == 0 {
		return
	}

	bestNum := numCrossingsBetweenLayers(segments, layers[idx-1], layers[idx])
	layer := make([]uint64, len(layers[idx]))

	for i := 0; i < o.Epochs; i++ {
		copy(layer, layers[idx])
		randomizeLayer(layer)

		numCrossings := numCrossingsBetweenLayers(segments, layers[idx-1], layer)

		// check if it is better then best
		if numCrossings < bestNum {
			bestNum = numCrossings
			copy(layers[idx], layer)
		}
	}
}

// MedianLayerOrderingOptimizer takes medium of upper (or lower) level neighbors for each node in layer.
// Median has property of stable vertical edges which is especially useful for "long" edges (fake nodes).
// Eades and Wormald, 1994
type MedianLayerOrderingOptimizer struct{}

func (o MedianLayerOrderingOptimizer) Optimize(_ Graph, _ LayeredGraph, _ int) {
	panic("TODO: implement this efficiently")
}

// time: O(N)
// space: O(N)
func randomizeLayer(layer []uint64) {
	n := len(layer)
	ordered := make([]uint64, n)
	for from, to := range rand.Perm(n) {
		ordered[to] = uint64(layer[from])
	}
	copy(layer, ordered)
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
