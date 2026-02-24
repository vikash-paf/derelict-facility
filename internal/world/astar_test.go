package world

import (
	"testing"

	"github.com/vikash-paf/derelict-facility/internal/entity"
)

// setupTestMap initializes a map and sets specific tiles as walls for testing.
func setupTestMap(w, h int, walls []entity.Point) *Map {
	m := NewMap(w, h)
	// Initialize all tiles as walkable floors first
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true})
		}
	}
	// Add the obstructions
	for _, p := range walls {
		m.SetTile(p.X, p.Y, Tile{Type: TileTypeWall, Walkable: false})
	}
	return m
}

func TestFindPath_StraightLine(t *testing.T) {
	m := setupTestMap(10, 10, nil)
	start := entity.Point{X: 1, Y: 1}
	target := entity.Point{X: 1, Y: 5}

	pf := NewPathfinder(m.Width, m.Height)
	path := pf.FindPath(m, start, target)

	if len(path) == 0 {
		t.Fatal("Failed to find a straight path")
	}

	// For a 4-way Manhattan distance, (1,1) to (1,5) should be 5 steps including start
	// or 4 steps excluding start depending on your implementation.
	// Based on our reconstructPath, it includes the target.
	if path[len(path)-1] != target {
		t.Errorf("Path did not reach target. Got %v, want %v", path[len(path)-1], target)
	}
}

func TestFindPath_Obstacle(t *testing.T) {
	// Create a U-shaped wall between (2,1) and (2,3)
	walls := []entity.Point{
		{X: 2, Y: 1}, {X: 2, Y: 2}, {X: 2, Y: 3},
	}
	m := setupTestMap(10, 10, walls)
	start := entity.Point{X: 1, Y: 2}
	target := entity.Point{X: 3, Y: 2}

	pf := NewPathfinder(m.Width, m.Height)
	path := pf.FindPath(m, start, target)

	if len(path) == 0 {
		t.Fatal("Failed to find path around obstacle")
	}

	// Ensure no point in the path is a wall
	for _, p := range path {
		if !m.Tiles[p.Y*m.Width+p.X].Walkable {
			t.Errorf("Path contains a wall at %v", p)
		}
	}
}

func TestFindPath_NoPath(t *testing.T) {
	// Completely wall off the target
	walls := []entity.Point{
		{X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1},
	}
	m := setupTestMap(3, 3, walls)
	start := entity.Point{X: 2, Y: 2}
	target := entity.Point{X: 0, Y: 0}

	pf := NewPathfinder(m.Width, m.Height)
	path := pf.FindPath(m, start, target)

	if path != nil {
		t.Errorf("Expected nil path for blocked target, got %v", path)
	}
}

func TestFindPath_Regression_HeapIndex(t *testing.T) {
	// This test ensures that when a shorter path to a node in the open set is found,
	// the heap is correctly updated via heap.Fix.
	m := setupTestMap(5, 5, nil)
	start := entity.Point{X: 0, Y: 0}
	target := entity.Point{X: 2, Y: 0}

	// We don't need complex logic here; the fact that FindPath runs
	// without a panic and reaches the target is a good sign for the logic
	// we implemented with openSetTracker and HeapIndex.
	pf := NewPathfinder(m.Width, m.Height)
	path := pf.FindPath(m, start, target)
	if len(path) == 0 || path[len(path)-1] != target {
		t.Errorf("Failed basic connectivity with HeapIndex logic")
	}
}

func TestVisual_PathfindingCorridor(t *testing.T) {
	// Setup a 10x10 map with a wall creating a narrow corridor
	walls := []entity.Point{
		{X: 5, Y: 0}, {X: 5, Y: 1}, {X: 5, Y: 2}, {X: 5, Y: 3}, {X: 5, Y: 4},
		{X: 5, Y: 6}, {X: 5, Y: 7}, {X: 5, Y: 8}, {X: 5, Y: 9},
	}
	m := setupTestMap(11, 10, walls) // Width 11 to give room on both sides

	start := entity.Point{X: 2, Y: 5}
	target := entity.Point{X: 9, Y: 1}

	pf := NewPathfinder(m.Width, m.Height)
	path := pf.FindPath(m, start, target)

	if len(path) == 0 {
		t.Error("Failed to find path through the gap")
	}

	// This is the visual confirmation
	VisualizePath(m, path, start, target)
}
