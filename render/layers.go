package render

import (
	"fmt"
	"math/rand"
	"sort"
)

// Layers is layered ordered graph.
//
// Example:
// 0: 1 8 11
// 1: 5 2
// 2: 11 2 3
type Layers [][]uint64

// NewLayers creates from partial ordering of layered ordered graph.
func NewLayers(nodeYX NodeYX) Layers {
	maxY := 0
	for _, yx := range nodeYX {
		if yx[0] > maxY {
			maxY = yx[0]
		}
	}

	layers := make([][]uint64, maxY+1)
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

// ToNodeYX returns partial order.
func (l Layers) ToNodeYX() NodeYX {
	nodeYX := map[uint64][2]int{}
	for y, layer := range l {
		for x, node := range layer {
			nodeYX[node] = [2]int{y, x}
		}
	}
	return nodeYX
}

// AssingRandomX will set random partial ordering.
func (l Layers) AssingRandomX() Layers {
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

// String makes pretty printed multiline string.
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
