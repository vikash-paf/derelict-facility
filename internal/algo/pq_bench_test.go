package algo

import (
	"container/heap"
	"math/rand"
	"testing"
)

// BenchmarkPriorityQueue_PushPop simulates a heavy pathfinding load
// where we are constantly adding and removing nodes.
func BenchmarkPriorityQueue_PushPop(b *testing.B) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	// Pre-fill the queue to simulate a mid-search state
	for i := 0; i < 1000; i++ {
		heap.Push(pq, &Node{
			FCost: rand.Intn(100),
			HCost: rand.Intn(50),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate exploring a neighbor
		heap.Push(pq, &Node{
			FCost: rand.Intn(100),
			HCost: rand.Intn(50),
		})
		// Simulate picking the next best node
		_ = heap.Pop(pq).(*Node)
	}
}

// when a better path is found (very common in A*).
func BenchmarkPriorityQueue_Fix(b *testing.B) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	var nodes []*Node
	for i := 0; i < 1000; i++ {
		n := &Node{FCost: 100 + i}
		nodes = append(nodes, n)
		heap.Push(pq, n)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Pick a random node and "find a better path"
		target := nodes[rand.Intn(1000)]
		target.FCost -= 10
		heap.Fix(pq, target.HeapIndex)
	}
}
