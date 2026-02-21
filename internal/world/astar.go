package world

import (
	"container/heap"

	"github.com/vikash-paf/derelict-facility/internal/algo"
	"github.com/vikash-paf/derelict-facility/internal/entity"
)

func FindPath(m *Map, start, target entity.Point) []entity.Point {
	// 1. Initial Validation
	targetTile := m.GetTile(target.X, target.Y)
	if targetTile == nil || !targetTile.Walkable {
		return nil
	}

	mapArea := m.Width * m.Height
	closedSet := make([]bool, mapArea)
	openSetTracker := make([]*algo.Node, mapArea)

	openSet := make(algo.PriorityQueue, 0)
	heap.Init(&openSet)

	// Start GCost must be 0 (the distance from start to start is zero)
	startNode := &algo.Node{
		Point: start,
		GCost: 0,
		HCost: ManhattanDistance(start, target),
	}
	startNode.FCost = startNode.GCost + startNode.HCost

	heap.Push(&openSet, startNode)
	openSetTracker[m.GetIndexFromPoint(start)] = startNode

	for openSet.Len() > 0 {
		currentNode := heap.Pop(&openSet).(*algo.Node)
		currIdx := m.GetIndexFromPoint(currentNode.Point)

		if currentNode.Point == target {
			return reconstructPath(currentNode)
		}

		openSetTracker[currIdx] = nil
		closedSet[currIdx] = true

		// Orthogonal neighbors (N, S, E, W)
		dx := []int{0, 0, 1, -1}
		dy := []int{-1, 1, 0, 0}

		for i := 0; i < 4; i++ {
			nx, ny := currentNode.Point.X+dx[i], currentNode.Point.Y+dy[i]

			// Boundary and Walkability check using your existing GetTile logic
			tile := m.GetTile(nx, ny)
			if tile == nil || !tile.Walkable {
				continue
			}

			nIdx := m.GetIndex(nx, ny)
			if closedSet[nIdx] {
				continue
			}

			newGCost := currentNode.GCost + 1
			neighborNode := openSetTracker[nIdx]

			if neighborNode == nil {
				// Discover new node
				newNode := &algo.Node{
					Point:  entity.Point{X: nx, Y: ny},
					Parent: currentNode,
					GCost:  newGCost,
					HCost:  ManhattanDistance(entity.Point{X: nx, Y: ny}, target),
				}
				newNode.FCost = newNode.GCost + newNode.HCost
				heap.Push(&openSet, newNode)
				openSetTracker[nIdx] = newNode
			} else if newGCost < neighborNode.GCost {
				// Found a more optimal path to a node already in the Open Set
				neighborNode.Parent = currentNode
				neighborNode.GCost = newGCost
				neighborNode.FCost = newGCost + neighborNode.HCost
				heap.Fix(&openSet, neighborNode.HeapIndex)
			}
		}
	}

	return nil
}
