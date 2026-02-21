package world

import (
	"strings"
	"testing"

	"github.com/vikash-paf/derelict-facility/internal/entity"
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
		wantVis    []entity.Point // Tiles that MUST be visible
		wantHidden []entity.Point // Tiles that MUST NOT be visible (behind walls)
	}{
		{
			name:       "blocked-horizontal",
			layoutName: "small_room",
			x0:         0, y0: 2, x1: 6, y1: 2,
			wantVis:    []entity.Point{{0, 2}, {3, 2}}, // Start and the Wall itself
			wantHidden: []entity.Point{{4, 2}, {5, 2}}, // Points behind the wall
		},
		{
			name:       "clear-path-vertical",
			layoutName: "small_room",
			x0:         1, y0: 0, x1: 1, y1: 4,
			wantVis:    []entity.Point{{1, 0}, {1, 2}, {1, 4}},
			wantHidden: []entity.Point{},
		},
		{
			name:       "long-range-obscured",
			layoutName: "large_corridor",
			x0:         2, y0: 2, x1: 20, y1: 2,
			wantVis:    []entity.Point{{2, 2}, {9, 2}},
			wantHidden: []entity.Point{{10, 2}, {15, 2}, {20, 2}},
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
				tt.name, tt.layoutName, tt.x0, tt.y0, tt.x1, tt.y1, InspectVisibility(m, 0, 0))

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

func TestComputeFOV(t *testing.T) {
	layouts := map[string]string{
		"pillar_room": `
...........
...........
.....#.....
...........
...........`,
		"closed_corridor": `
###########
#.........#
#....#....#
#.........#
###########`,
	}

	tests := []struct {
		name       string
		layoutName string
		playerX    int
		playerY    int
		radius     int
		// Points we want to verify
		mustSee    []entity.Point
		mustNotSee []entity.Point
	}{
		{
			name:       "short-range-no-obstruction",
			layoutName: "pillar_room",
			playerX:    5, playerY: 0,
			radius:     2,
			mustSee:    []entity.Point{{5, 0}, {5, 2}}, // Directly below
			mustNotSee: []entity.Point{{5, 4}},         // Too far for radius 2
		},
		{
			name:       "shadow-behind-pillar",
			layoutName: "pillar_room",
			playerX:    5, playerY: 1,
			radius:     4,
			mustSee:    []entity.Point{{5, 2}}, // The wall itself
			mustNotSee: []entity.Point{{5, 3}}, // Shadow cast by wall at (5,2)
		},
		{
			name:       "trapped-in-box",
			layoutName: "closed_corridor",
			playerX:    5, playerY: 2,
			radius:     10,
			mustSee:    []entity.Point{{1, 2}, {9, 2}},  // Internal walls
			mustNotSee: []entity.Point{{0, 2}, {10, 2}}, // Outside the map boundary walls
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestMap(layouts[tt.layoutName])

			m.ComputeFOV(tt.playerX, tt.playerY, tt.radius)

			// Visual Debug Output
			t.Logf("\nTest: %s\nPlayer at (%d,%d), Radius: %d\n%s",
				tt.name, tt.playerX, tt.playerY, tt.radius, InspectVisibility(m, tt.playerX, tt.playerY))

			// Assertions
			for _, p := range tt.mustSee {
				if !m.Tiles[p.X+p.Y*m.Width].Visible {
					t.Errorf("Expected point {%d, %d} to be visible", p.X, p.Y)
				}
			}
			for _, p := range tt.mustNotSee {
				if m.Tiles[p.X+p.Y*m.Width].Visible {
					t.Errorf("Expected point {%d, %d} to be hidden", p.X, p.Y)
				}
			}
		})
	}
}
