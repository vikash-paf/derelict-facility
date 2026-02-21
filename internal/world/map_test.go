package world

import (
	"fmt"
	"strings"
	"testing"
)

// Helper to create a map from a string for easy testing
// # = Wall (Not Walkable)
// . = Floor (Walkable)
func newTestMap(layout string) *Map {
	lines := strings.Split(strings.TrimSpace(layout), "\n")
	height := len(lines)
	width := len(strings.TrimSpace(lines[0]))
	m := &Map{
		Width:  width,
		Height: height,
		Tiles:  make([]Tile, width*height),
	}

	for y, line := range lines {
		for x, char := range strings.TrimSpace(line) {
			idx := x + y*width
			if char == '#' {
				m.Tiles[idx] = Tile{Type: TileTypeWall, Walkable: false}
			} else {
				m.Tiles[idx] = Tile{Type: TileTypeFloor, Walkable: true}
			}
		}
	}
	return m
}

func TestCastRay_WallStopping(t *testing.T) {
	// 1. Setup a map with a wall at (3, 2)
	// Grid is 7x5
	layout := `
.......
.......
...#...
.......
.......
`
	m := newTestMap(layout)

	// 2. Cast a ray from left to right, passing through the wall
	startX, startY := 0, 2
	endX, endY := 6, 2
	m.castRay(startX, startY, endX, endY)

	// 3. Log the visual output so you can see it in 'go test -v'
	fmt.Printf("Ray from (%d,%d) to (%d,%d):\n%s", startX, startY, endX, endY, InspectVisibility(m))

	// 4. Verification Logic
	// Tile (3,2) should be Visible (the ray hit it)
	if !m.Tiles[3+2*m.Width].Visible {
		t.Errorf("Wall at (3,2) should be visible (hit by ray)")
	}

	// Tile (4,2) should NOT be visible (it's behind the wall)
	if m.Tiles[4+2*m.Width].Visible {
		t.Errorf("Tile at (4,2) should be hidden behind the wall")
	}

	// Tile (0,2) should be visible (origin)
	if !m.Tiles[0+2*m.Width].Visible {
		t.Error("Starting tile should be visible")
	}
}
