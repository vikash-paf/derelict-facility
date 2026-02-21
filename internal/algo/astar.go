package algo

import (
	"container/heap"

	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/math"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// ManhattanDistance calculates the heuristic (H-Cost) without diagonal movement.
func ManhattanDistance(p1, p2 entity.Point) int {
	return math.Abs(p1.X-p2.X) + math.Abs(p1.Y-p2.Y)
}

// FindPath returns list of points representing the shortest path.
func FindPath(m *world.Map, start, target entity.Point) []entity.Point {
	// 1. Setup the Open Set (tiles to evaluate) and Closed Set (tiles already evaluated)
	openSet := make(PriorityQueue, 0)
	heap.Init(&openSet)

	// We use a map for the closed set for O(1) lookups. Key: "x,y" string or a 1D index.
	// For performance, a 1D index (y * width + x) is best!
	// closedSet := make(map[int]bool)

	// 2. Add the start node to the open set
	startNode := &Node{
		Point: start,
		GCost: 0,
		HCost: ManhattanDistance(start, target),
	}
	startNode.FCost = startNode.GCost + startNode.HCost
	heap.Push(&openSet, startNode)

	// Keep track of nodes in the open set to update them if we find a faster route
	openSetTracker := make(map[int]*Node)
	openSetTracker[start.Y*m.Width+start.X] = startNode

	// 3. The Search Loop
	for openSet.Len() > 0 {
		// TODO: Pop the node with the lowest FCost from openSet using:
		// currentNode := heap.Pop(&openSet).(*Node)

		// TODO: Remove currentNode from openSetTracker

		// TODO: Calculate 1D index of currentNode and add it to closedSet

		// TODO: Are we at the target?
		// If currentNode.Point == target:
		// We won! Reconstruct the path by following currentNode.Parent backward,
		// reverse the slice, and return it.

		// TODO: Get neighbors (Up, Down, Left, Right).
		// For each neighbor:
		//   1. Check if it's outside the map or Walkable == false. If so, ignore.
		//   2. Calculate its 1D index. Is it in the closedSet? If so, ignore.
		//   3. Calculate new GCost = currentNode.GCost + 1
		//   4. Check if neighbor is in openSetTracker.
		//      - If not: Create a new Node, set its Parent to currentNode, calculate G, H, F.
		//                Push it to openSet and add to openSetTracker.
		//      - If it IS in openSetTracker, but this new GCost is LOWER than its current GCost:
		//                Update its GCost, FCost, and Parent.
		//                Call heap.Fix(&openSet, neighborNode.HeapIndex) to re-sort the queue.
	}

	// No path found!
	return nil
}
