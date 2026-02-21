package algo

import (
	"container/heap"
	"testing"

	"github.com/vikash-paf/derelict-facility/internal/entity"
)

func TestPriorityQueueOrdering(t *testing.T) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	nodes := []*Node{
		{Point: entity.Point{X: 1, Y: 1}, FCost: 10, HCost: 5},
		{Point: entity.Point{X: 2, Y: 2}, FCost: 5, HCost: 2},
		{Point: entity.Point{X: 3, Y: 3}, FCost: 15, HCost: 10},
		{Point: entity.Point{X: 4, Y: 4}, FCost: 5, HCost: 1}, // Lower H-Cost tie-breaker
	}

	for _, n := range nodes {
		heap.Push(pq, n)
	}

	// First element should be (4,4) because FCost is 5 and HCost is 1 (tie-breaker)
	first := heap.Pop(pq).(*Node)
	if first.Point.X != 4 || first.FCost != 5 {
		t.Errorf("Expected node (4,4) with FCost 5, got (%d,%d) with FCost %d",
			first.Point.X, first.Point.Y, first.FCost)
	}

	// Second element should be (2,2) because FCost is 5
	second := heap.Pop(pq).(*Node)
	if second.Point.X != 2 || second.FCost != 5 {
		t.Errorf("Expected node (2,2) with FCost 5, got (%d,%d)", second.Point.X, second.Point.Y)
	}

	// Third should be (1,1) with FCost 10
	third := heap.Pop(pq).(*Node)
	if third.FCost != 10 {
		t.Errorf("Expected FCost 10, got %d", third.FCost)
	}
}

// This is the most important stuff for the performance of A* when we find a better path to a node already in the PQ.
func TestHeapIndexConsistency(t *testing.T) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	nodes := []*Node{
		{Point: entity.Point{X: 0, Y: 0}, FCost: 100},
		{Point: entity.Point{X: 1, Y: 1}, FCost: 50},
		{Point: entity.Point{X: 2, Y: 2}, FCost: 25},
	}

	for _, n := range nodes {
		heap.Push(pq, n)
	}

	// Check if every node knows its own position in the underlying slice
	for i, node := range *pq {
		if node.HeapIndex != i {
			t.Errorf("Node at index %d has incorrect HeapIndex %d", i, node.HeapIndex)
		}
	}

	// Manually trigger a fix or a pop and re-verify
	heap.Pop(pq)
	for i, node := range *pq {
		if node.HeapIndex != i {
			t.Errorf("After Pop, node at index %d has incorrect HeapIndex %d", i, node.HeapIndex)
		}
	}
}

func TestEmptyQueue(t *testing.T) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	if pq.Len() != 0 {
		t.Errorf("Expected empty queue length 0, got %d", pq.Len())
	}
}
