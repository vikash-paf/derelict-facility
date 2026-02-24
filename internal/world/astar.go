package world

import (
	"container/heap"

	"github.com/vikash-paf/derelict-facility/internal/algo"
	"github.com/vikash-paf/derelict-facility/internal/entity"
)

// Pathfinder holds reusable buffers for A* pathfinding to avoid allocations.
type Pathfinder struct {
	closedSet      []uint64
	openSetTracker []*algo.Node
	openSet        algo.PriorityQueue
	generation     uint64
}

// NewPathfinder initializes the buffers for a map of the given dimensions.
func NewPathfinder(width, height int) *Pathfinder {
	mapArea := width * height
	return &Pathfinder{
		closedSet:      make([]uint64, mapArea),
		openSetTracker: make([]*algo.Node, mapArea),
		openSet:        make(algo.PriorityQueue, 0, 100),
		generation:     0,
	}
}

func (pf *Pathfinder) FindPath(m *Map, start, target entity.Point) []entity.Point {
	// 1. Initial Validation
	targetTile := m.GetTile(target.X, target.Y)
	if targetTile == nil || !targetTile.Walkable {
		return nil
	}

	pf.generation++
	pf.openSet = pf.openSet[:0] // Clear queue without reallocating
	clear(pf.openSetTracker)    // Quickly nil out all pointers

	// Start GCost must be 0 (the distance from start to start is zero)
	startNode := &algo.Node{
		Point: start,
		GCost: 0,
		HCost: ManhattanDistance(start, target),
	}
	startNode.FCost = startNode.GCost + startNode.HCost

	heap.Push(&pf.openSet, startNode)
	pf.openSetTracker[m.GetIndexFromPoint(start)] = startNode

	for pf.openSet.Len() > 0 {
		currentNode := heap.Pop(&pf.openSet).(*algo.Node)
		currIdx := m.GetIndexFromPoint(currentNode.Point)

		if currentNode.Point == target {
			return reconstructPath(currentNode)
		}

		pf.openSetTracker[currIdx] = nil
		pf.closedSet[currIdx] = pf.generation

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
			if pf.closedSet[nIdx] == pf.generation {
				continue
			}

			newGCost := currentNode.GCost + 1
			neighborNode := pf.openSetTracker[nIdx]

			if neighborNode == nil {
				// Discover new node
				newNode := &algo.Node{
					Point:  entity.Point{X: nx, Y: ny},
					Parent: currentNode,
					GCost:  newGCost,
					HCost:  ManhattanDistance(entity.Point{X: nx, Y: ny}, target),
				}
				newNode.FCost = newNode.GCost + newNode.HCost
				heap.Push(&pf.openSet, newNode)
				pf.openSetTracker[nIdx] = newNode
			} else if newGCost < neighborNode.GCost {
				// Found a more optimal path to a node already in the Open Set
				neighborNode.Parent = currentNode
				neighborNode.GCost = newGCost
				neighborNode.FCost = newGCost + neighborNode.HCost
				heap.Fix(&pf.openSet, neighborNode.HeapIndex)
			}
		}
	}

	return nil
}
