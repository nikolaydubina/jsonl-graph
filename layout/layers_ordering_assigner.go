package layout

import (
	"math/rand"
	"sort"
)

type LayerOrderingOptimizer interface {
	Optimize(g Graph, lg LayeredGraph, layer int)
}

// LBLOrderingOptimizer is layer-by-layer sweep ordering optimizer.
// Goes up and down number of iterations across all layers.
// Considers upper and lower fixed and permutes ordering in layer.
type LBLOrderingOptimizer struct {
	Epochs                 int
	LayerOrderingOptimizer LayerOrderingOptimizer
}

func (o LBLOrderingOptimizer) Optimize(g Graph, lg LayeredGraph) {
	for i := range lg.Layers() {
		RandomLayerOrderingAssigner{}.Assign(g, lg, i)
	}

	bestNumCrossings := lg.NumCrossings()
	bestNodeYX := make(NodeYX, len(lg.NodeYX))
	for node, yx := range lg.NodeYX {
		bestNodeYX[node] = yx
	}

	for t := 0; t < o.Epochs; t++ {
		layers := lg.Layers()
		for i := range layers {
			j := i
			if (i % 2) == 0 {
				j = len(layers) - 1 - i
			}
			o.LayerOrderingOptimizer.Optimize(g, lg, j)
		}

		if numCrossings := lg.NumCrossings(); numCrossings < bestNumCrossings {
			bestNumCrossings = numCrossings
			for node, yx := range lg.NodeYX {
				bestNodeYX[node] = yx
			}
		}
	}

	for node, yx := range bestNodeYX {
		lg.NodeYX[node] = yx
	}
}

// RandomLayerOrderingOptimizer picks best out of epochs random orderings.
type RandomLayerOrderingOptimizer struct {
	Epochs int
}

func (o RandomLayerOrderingOptimizer) Optimize(_ Graph, lg LayeredGraph, layerIdx int) {
	bestNumCrossings := lg.NumCrossingsAtLayer(layerIdx)
	bestNodeYX := make(NodeYX, len(lg.NodeYX))
	for node, yx := range lg.NodeYX {
		bestNodeYX[node] = yx
	}

	for i := 0; i < o.Epochs; i++ {
		layers := lg.Layers()
		layer := make([]uint64, len(layers[layerIdx]))

		// permute
		for iold, inew := range rand.Perm(len(layer)) {
			layer[inew] = layers[layerIdx][iold]
		}

		// update layers data
		for i, node := range layer {
			lg.NodeYX[node] = [2]int{layerIdx, i}
		}

		// check if it is better then best
		if numCrossings := lg.NumCrossingsAtLayer(layerIdx); numCrossings < bestNumCrossings {
			bestNumCrossings = numCrossings
			for node, yx := range lg.NodeYX {
				bestNodeYX[node] = yx
			}
		}
	}

	for node, yx := range bestNodeYX {
		lg.NodeYX[node] = yx
	}
}

// MedianLayerOrderingOptimizer takes medium of upper (or lower) level neighbors for each node in layer.
// Median has property of stable vertical edges which is especially useful for "long" edges (fake nodes).
// Eades and Wormald, 1994
type MedianLayerOrderingOptimizer struct {
	AlignLowerLevel bool
}

func (o MedianLayerOrderingOptimizer) Optimize(_ Graph, lg LayeredGraph, layerIdx int) {
	var layer []uint64

	for node, yx := range lg.NodeYX {
		// first layer does not have higher layer
		if o.AlignLowerLevel && yx[0] == 0 {
			continue
		}
		// bottom layer does not have lower layer
		if !o.AlignLowerLevel && yx[0] == len(lg.Layers()) {
			continue
		}

		if yx[0] != layerIdx {
			continue
		}

		// median of target level connected nodes
		var targets []uint64
		for e := range lg.Segments {
			target := e[0]
			if o.AlignLowerLevel {
				target = e[1]
			}
			if target == node {
				targets = append(targets, target)
			}
		}

		if len(targets) > 0 {
			median := targets[len(targets)/2]
			lg.NodeYX[node] = [2]int{yx[0], lg.NodeYX[median][1]}
		}
	}

	sort.Slice(layer, func(i, j int) bool { return layer[i] < layer[j] })

	for i, node := range layer {
		lg.NodeYX[node] = [2]int{layerIdx, i}
	}
}

// RandomLayerOrderingAssigner will assign random permutation to nodes in layer.
type RandomLayerOrderingAssigner struct{}

func (g RandomLayerOrderingAssigner) Assign(_ Graph, lg LayeredGraph, layerIdx int) {
	layers := lg.Layers()

	n := len(layers[layerIdx])
	ordered := make([]uint64, n)
	for from, to := range rand.Perm(n) {
		ordered[to] = uint64(layers[layerIdx][from])
	}
	copy(layers[layerIdx], ordered)

	for y, layer := range layers {
		for x, node := range layer {
			lg.NodeYX[node] = [2]int{y, x}
		}
	}
}
