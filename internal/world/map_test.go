package world

import (
	"strings"
	"testing"
)

// Helper to create a map from a string for easy testing
// # = Wall (Not Walkable)
// . = Floor (Walkable)
func newTestMap(layout string) *Map {
	// Trim only the leading/trailing newlines of the whole block
	layout = strings.Trim(layout, "\n")
	lines := strings.Split(layout, "\n")

	height := len(lines)
	width := len(lines[0]) // Assume all lines are the same length

	m := &Map{
		Width:  width,
		Height: height,
		Tiles:  make([]Tile, width*height),
	}

	for y, line := range lines {
		for x, char := range line {
			if x >= width {
				break
			} // Safety check for uneven strings

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

func TestCastRay(t *testing.T) {
	layouts := map[string]string{
		"small_room": `
.......
.......
...#...
.......
.......`,
		"large_corridor": `
......................
......................
.........####.........
......................
......................
......................`,
	}

	tests := []struct {
		name       string
		layoutName string
		x0, y0     int
		x1, y1     int
		wantVis    []Point // Tiles that MUST be visible
		wantHidden []Point // Tiles that MUST NOT be visible (behind walls)
	}{
		{
			name:       "blocked-horizontal",
			layoutName: "small_room",
			x0:         0, y0: 2, x1: 6, y1: 2,
			wantVis:    []Point{{0, 2}, {3, 2}}, // Start and the Wall itself
			wantHidden: []Point{{4, 2}, {5, 2}}, // Points behind the wall
		},
		{
			name:       "clear-path-vertical",
			layoutName: "small_room",
			x0:         1, y0: 0, x1: 1, y1: 4,
			wantVis:    []Point{{1, 0}, {1, 2}, {1, 4}},
			wantHidden: []Point{},
		},
		{
			name:       "long-range-obscured",
			layoutName: "large_corridor",
			x0:         2, y0: 2, x1: 20, y1: 2,
			wantVis:    []Point{{2, 2}, {9, 2}},
			wantHidden: []Point{{10, 2}, {15, 2}, {20, 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize map from our library
			layout, ok := layouts[tt.layoutName]
			if !ok {
				t.Fatalf("Layout %s not found", tt.layoutName)
			}
			m := newTestMap(layout)

			// Fire the ray
			m.castRay(tt.x0, tt.y0, tt.x1, tt.y1)

			// Log the visual state for inspection on failure or -v
			t.Logf("\nTest: %s (Map: %s)\nRay: (%d,%d) -> (%d,%d)\n%s",
				tt.name, tt.layoutName, tt.x0, tt.y0, tt.x1, tt.y1, InspectVisibility(m))

			// Check expected visible points
			for _, p := range tt.wantVis {
				idx := p.X + p.Y*m.Width
				if !m.Tiles[idx].Visible {
					t.Errorf("Expected point {%d, %d} to be visible, but it was hidden", p.X, p.Y)
				}
			}

			// Check expected hidden points
			for _, p := range tt.wantHidden {
				idx := p.X + p.Y*m.Width
				if m.Tiles[idx].Visible {
					t.Errorf("Expected point {%d, %d} to be hidden, but it was visible", p.X, p.Y)
				}
			}
		})
	}
}
