package algo

import (
	"github.com/vikash-paf/derelict-facility/internal/entity"
)

// Node represents a single tile on the grid during our A* search.
type Node struct {
	Point     entity.Point // The X, Y coordinates
	Parent    *Node        // The tile we came from (used to trace the path back)
	GCost     int          // Distance from the start
	HCost     int          // Heuristic (estimated distance to the target)
	FCost     int          // G + H
	HeapIndex int          // Required by container/heap to manage the array
}

// PriorityQueue implements heap.Interface and holds Nodes.
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

// Less ensures we pop the node with the lowest F-Cost.
func (pq PriorityQueue) Less(i, j int) bool {
	// If F-Costs are equal, tie-break by H-Cost (closest to target)
	if pq[i].FCost == pq[j].FCost {
		return pq[i].HCost < pq[j].HCost
	}
	return pq[i].FCost < pq[j].FCost
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].HeapIndex = i
	pq[j].HeapIndex = j
}

// Push and Pop use empty interfaces, which is just how Go < 1.18 did generics.
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	node := x.(*Node)
	node.HeapIndex = n
	*pq = append(*pq, node)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil      // Avoid memory leak
	node.HeapIndex = -1 // For safety
	*pq = old[0 : n-1]
	return node
}
