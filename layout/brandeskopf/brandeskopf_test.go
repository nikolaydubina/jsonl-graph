package brandeskopf

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fake node ID
func fake(from, to uint64, i int) uint64 {
	return 100000 + from*100 + to*10 + uint64(i)
}

func makeLongEdge(from, to uint64, n int) [][2]uint64 {
	segments := make([][2]uint64, n+1)
	segments[0] = [2]uint64{from, fake(from, to, 1)}
	for i := 1; i < n; i++ {
		segments[i] = [2]uint64{fake(from, to, i), fake(from, to, i+1)}
	}
	segments[n] = [2]uint64{fake(from, to, n), to}
	return segments
}

func toNodeYX(layers [][]uint64) map[uint64][2]int {
	nodeYX := map[uint64][2]int{}
	for y, layer := range layers {
		for x, node := range layer {
			nodeYX[node] = [2]int{y, x}
		}
	}
	return nodeYX
}

func sortuint64(v []uint64) []uint64 {
	sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
	return v
}

func TestReferenceGraphFromPaper(t *testing.T) {
	layers := [][]uint64{
		{1, 2},
		{fake(1, 13, 1), fake(1, 21, 1), fake(1, 4, 1), 3, fake(2, 20, 1)},
		{fake(1, 13, 2), fake(1, 21, 2), 4, 5, fake(3, 23, 1), fake(2, 20, 2)},
		{fake(1, 13, 3), fake(1, 21, 3), 6, 7, fake(3, 23, 2), fake(2, 20, 3)},
		{fake(1, 13, 4), fake(1, 21, 4), 8, fake(1, 16, 1), fake(6, 23, 1), 9, fake(3, 23, 3), fake(2, 20, 4)},
		{fake(1, 13, 5), fake(1, 21, 5), 10, 11, fake(6, 23, 2), fake(6, 23, 2), 12, fake(3, 23, 4), fake(2, 20, 5)},
		{13, fake(1, 21, 6), 14, 15, 16, fake(6, 23, 3), fake(12, 20, 1), fake(3, 23, 5), fake(2, 20, 6)},
		{17, fake(1, 21, 7), 18, 19, fake(6, 23, 4), 20, fake(3, 23, 6)},
		{21, 22, fake(6, 23, 5), fake(3, 23, 7)},
		{23},
	}

	dummy := map[uint64]bool{}
	for _, layer := range layers {
		for _, node := range layer {
			if node > 100 {
				dummy[node] = true
			}
		}
	}

	edgeList := [][2]uint64{
		{1, 13},
		{1, 21},
		{1, 4},
		{1, 3},
		{2, 3},
		{2, 20},
		{3, 4},
		{3, 5},
		{3, 23},
		{4, 6},
		{5, 7},
		{6, 8},
		{6, 16},
		{6, 23},
		{7, 9},
		{8, 10},
		{8, 11},
		{9, 12},
		{10, 13},
		{10, 14},
		{10, 15},
		{11, 15},
		{11, 16},
		{12, 20},
		{13, 17},
		{14, 17},
		{14, 18},
		{16, 18},
		{16, 19},
		{16, 20},
		{18, 21},
		{19, 22},
		{21, 23},
		{22, 23},
	}

	var longEdgesList [][2]uint64
	longEdgesList = append(longEdgesList, makeLongEdge(1, 13, 5)...)
	longEdgesList = append(longEdgesList, makeLongEdge(1, 21, 7)...)
	longEdgesList = append(longEdgesList, makeLongEdge(1, 4, 1)...)
	longEdgesList = append(longEdgesList, makeLongEdge(2, 20, 6)...)
	longEdgesList = append(longEdgesList, makeLongEdge(3, 20, 5)...)
	longEdgesList = append(longEdgesList, makeLongEdge(6, 16, 2)...)
	longEdgesList = append(longEdgesList, makeLongEdge(6, 23, 5)...)
	longEdgesList = append(longEdgesList, makeLongEdge(12, 20, 1)...)

	segments := map[[2]uint64]bool{}
	for _, e := range longEdgesList {
		segments[e] = true
	}
	for _, e := range edgeList {
		segments[e] = true
	}

	g := LayeredGraph{
		Segments: segments,
		Dummy:    dummy,
		Layers:   layers,
		NodeYX:   toNodeYX(layers),
	}

	t.Run("check graph is correct", func(t *testing.T) {
		assert.Equal(t, 10, len(layers))

		numNodes := 0
		numFakes := 0
		for _, layer := range layers {
			for _, node := range layer {
				if node > 1000 {
					numFakes++
				} else {
					numNodes++
				}
			}
		}

		assert.Equal(t, 23, numNodes)
		assert.Equal(t, 34, numFakes)
	})

	t.Run("check upper neighbors", func(t *testing.T) {
		assert.Equal(t, []uint64{1, 2}, sortuint64(g.UpperNeighbors(3)))
		assert.Equal(t, []uint64{14, 16}, sortuint64(g.UpperNeighbors(18)))
		assert.Equal(t, []uint64{3, fake(1, 4, 1)}, sortuint64(g.UpperNeighbors(4)))
	})

	t.Run("check Alg 1, preprocessing", func(t *testing.T) {
		typeOneSegments := preprocessing(g)
		t.Logf("%#v", typeOneSegments)
	})
}
