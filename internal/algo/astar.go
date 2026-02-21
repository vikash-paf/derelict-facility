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

func FindPath(m *world.Map, start, target entity.Point) []entity.Point {
	// Total number of tiles
	mapArea := m.Width * m.Height

	// Replace maps with pre-allocated slices for O(1) access without hashing
	closedSet := make([]bool, mapArea)
	openSetTracker := make([]*Node, mapArea)

	openSet := make(PriorityQueue, 0)
	heap.Init(&openSet)

	startNode := &Node{
		Point: start,
		GCost: 0,
		HCost: ManhattanDistance(start, target),
	}
	startNode.FCost = startNode.GCost + startNode.HCost

	heap.Push(&openSet, startNode)
	openSetTracker[start.Y*m.Width+start.X] = startNode

	for openSet.Len() > 0 {
		currentNode := heap.Pop(&openSet).(*Node)
		currIdx := currentNode.Point.Y*m.Width + currentNode.Point.X

		// Mark as nil in tracker since it's no longer "Open"
		openSetTracker[currIdx] = nil

		if currentNode.Point == target {
			return reconstructPath(currentNode)
		}

		closedSet[currIdx] = true

		// Neighbors: N, S, W, E
		dx := []int{0, 0, -1, 1}
		dy := []int{-1, 1, 0, 0}

		for i := 0; i < 4; i++ {
			nx, ny := currentNode.Point.X+dx[i], currentNode.Point.Y+dy[i]

			// Boundary and Walkability check
			if nx < 0 || ny < 0 || nx >= m.Width || ny >= m.Height {
				continue
			}

			nIdx := ny*m.Width + nx
			if closedSet[nIdx] || !m.Tiles[nIdx].Walkable {
				continue
			}

			newGCost := currentNode.GCost + 1
			neighborNode := openSetTracker[nIdx]

			if neighborNode == nil {
				// New Node
				newNode := &Node{
					Point:  entity.Point{X: nx, Y: ny},
					Parent: currentNode,
					GCost:  newGCost,
					HCost:  ManhattanDistance(entity.Point{X: nx, Y: ny}, target),
				}
				newNode.FCost = newNode.GCost + newNode.HCost
				heap.Push(&openSet, newNode)
				openSetTracker[nIdx] = newNode
			} else if newGCost < neighborNode.GCost {
				// Improved path
				neighborNode.Parent = currentNode
				neighborNode.GCost = newGCost
				neighborNode.FCost = newGCost + neighborNode.HCost
				heap.Fix(&openSet, neighborNode.HeapIndex)
			}
		}
	}
	return nil
}

// reconstructPath follows the parent pointers back to the start.
func reconstructPath(endNode *Node) []entity.Point {
	var path []entity.Point
	curr := endNode
	for curr != nil {
		path = append(path, curr.Point)
		curr = curr.Parent
	}
	// The path is currently [target -> start], we need [start -> target]
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
