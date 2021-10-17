package brandeskopf

import (
	"log"
	"math"
)

// LayeredGraph is graph with Dummy nodes such that there is no long edges.
// Short edge is between nodes in Layers next to each other.
// Long edge is between nodes in 1+ Layers between each other.
// Segments are short edges and long edges.
// Top layer has lowest layer number.
type LayeredGraph struct {
	Segments map[[2]uint64]bool // segment is an edge in layered graph, can be real edge or piece of fake edge
	Dummy    map[uint64]bool    // fake nodes
	NodeYX   map[uint64][2]int  // node -> {layer, ordering in layer}
	Layers   [][]uint64         // same as NodeYX but different form
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

// "Fast and Simple Horizontal Coordinate Assignment" by Ulrik Brandes and Boris Kopf, 2002
// Computes horizontal coordinate in layered graph, given ordering within each layer.
// Produces result such that neighbors are close and long edges cross Layers are straight.
// Works on fully connected graphs.
// delta is minimum horizontal separation between nodes.
// This is Alg 4 in paper.
// Assuming nodes do not have width.
func BrandesKopfLayersHorizontalAssignment(g LayeredGraph, delta int) (x map[uint64]int) {
	x = make(map[uint64]int, len(g.NodeYX))

	typeOneSegments := preprocessing(g)
	for _, vdirection := range []int{-1, 1} {
		for _, hdirection := range []int{-1, 1} {
			log.Printf("%v %v", vdirection, hdirection)
			root, align := verticalAlignment(g, typeOneSegments)
			log.Printf("brandes kopf: vertical done")
			x = horizontalCompaction(g, root, align, delta)
			log.Printf("brandes kopf: horizontal done")
		}
	}

	// TODO: balancing by taking median for every node.
	return x
}

// Alg 1.
// Type 1 conflicts arise when a non-inner segment (normal edge) crosses an inner segment (edge between two fake nodes).
// The algorithm traverses Layers from left to right (index l) while maintaining the upper neighbors, v(i)_k0 and v(i)_k1, of the two closest inner Segments.
func preprocessing(g LayeredGraph) (typeOneSegments map[[2]uint64]bool) {
	typeOneSegments = map[[2]uint64]bool{}

	for i := range g.Layers {
		if i == (len(g.Layers) - 1) {
			continue
		}
		nextLayer := g.Layers[i+1]

		k0 := 0
		l := 0

		for l1, v := range nextLayer {
			var upperNeighborInnerSegment uint64
			for _, u := range g.UpperNeighbors(v) {
				if g.IsInnerSegment([2]uint64{u, v}) {
					upperNeighborInnerSegment = u
					break
				}
			}

			if (l1 == (len(nextLayer) - 1)) || upperNeighborInnerSegment != 0 {
				k1 := len(g.Layers[i]) - 1
				if upperNeighborInnerSegment != 0 {
					k1 = g.NodeYX[upperNeighborInnerSegment][1]
				}
				for l <= l1 {
					for k, u := range g.UpperNeighbors(nextLayer[l]) {
						if k < k0 || k > k1 {
							typeOneSegments[[2]uint64{u, v}] = true
						}
					}
					l += 1
				}
				k0 = k1
			}
		}
	}

	return typeOneSegments
}

// Alg 2.
// Obtain a leftmost alignment with upper neighbors.
// A maximal set of vertically aligned vertices is called a block, and we define the root of a block to be its topmost vertex.
// Blocks are stored as cyclicly linked lists, each node has reference to its lower aligned neighbor and lowest refers to topmost.
// Each node has additional reference to root of its block.
func verticalAlignment(g LayeredGraph, typeOneSegments map[[2]uint64]bool) (root map[uint64]uint64, align map[uint64]uint64) {
	root = make(map[uint64]uint64, len(g.NodeYX))
	align = make(map[uint64]uint64, len(g.NodeYX))

	for v := range g.NodeYX {
		root[v] = v
		align[v] = v
	}

	for i := range g.Layers {
		r := 0
		for _, v := range g.Layers[i] {
			upNeighbors := g.UpperNeighbors(v)
			if d := len(upNeighbors); d > 0 {
				for m := d / 2; m < ((d+1)/2) && (m < d); m++ {
					u := upNeighbors[m]
					if align[v] == v {
						if !typeOneSegments[[2]uint64{u, v}] && r < g.NodeYX[u][1] {
							align[u] = v
							root[v] = root[u]
							align[v] = root[v]
							r = g.NodeYX[u][1]
						}
					}
				}
			}
		}
	}

	return root, align
}

// part of Alg 3.
func placeBlock(g LayeredGraph, x map[uint64]int, root map[uint64]uint64, align map[uint64]uint64, sink map[uint64]uint64, shift map[uint64]int, delta int, v uint64) {
	if _, ok := x[v]; !ok {
		x[v] = 0
		flag := true
		for w := v; flag; flag = v != w {
			if g.NodeYX[w][1] > 0 {
				u := root[g.Layers[g.NodeYX[w][0]][g.NodeYX[w][1]-1]]
				placeBlock(g, x, root, align, sink, shift, delta, u)
				if sink[v] == v {
					sink[v] = sink[u]
				}
				if sink[v] != sink[u] {
					if s := x[v] - x[u] - delta; s < shift[sink[u]] {
						shift[sink[u]] = s
					}
				} else {
					if s := x[u] + delta; s > x[v] {
						x[v] = s
					}
				}
			}
			w = align[w]
		}
	}
}

// Alg 3.
// All node of a block are assigned the coordinate of the root.
// Partition each block in to classes.
// Class is defined by reachable sink which has the topmost root
// Within each class, we apply a longest path layering,
// i.e. the relative coordinate of a block with respect to the defining
// sink is recursively determined to be the maximum coordinate of
// the preceding blocks in the same class, plus minimum separation.
// For each class, from top to bottom, we then compute the absolute coordinates
// of its members by placing the class with minimum separation from previously placed classes.
func horizontalCompaction(g LayeredGraph, root map[uint64]uint64, align map[uint64]uint64, delta int) (x map[uint64]int) {
	sink := map[uint64]uint64{}
	shift := map[uint64]int{}
	x = map[uint64]int{}

	for v := range g.NodeYX {
		sink[v] = v
		shift[v] = math.MaxInt
	}

	// root coordinates relative to sink
	for v := range g.NodeYX {
		if root[v] == v {
			placeBlock(g, x, root, align, sink, shift, delta, v)
		}
	}

	// absolute coordinates
	for v := range g.NodeYX {
		x[v] = x[root[v]]
		if s := shift[sink[root[v]]]; s < math.MaxInt {
			x[v] += s
		}
	}

	return x
}
