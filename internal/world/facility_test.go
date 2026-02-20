package world

import (
	"testing"
)

func TestFacilityGenerator_Generate(t *testing.T) {
	countTiles := func(m *Map) (walls, walkable int) {
		for x := 0; x < m.Width; x++ {
			for y := 0; y < m.Height; y++ {
				tile := m.GetTile(x, y)
				if tile.Type == TileTypeWall && !tile.Walkable {
					walls++
				} else if tile.Walkable {
					walkable++
				}
			}
		}
		return
	}

	tests := []struct {
		name          string
		width, height int
		expectNilMap  bool
	}{
		{"Zero dimensions", 0, 0, true},
		{"Negative dimensions", -5, -5, true},
		{"Minimal valid map", roomMinSize, roomMinSize, false},
		{"Small square map", 4, 4, false},
		{"Wide map with minimal height", 20, roomMinSize, false},
		{"Tall map with minimal width", roomMinSize, 30, false},
		{"Rectangular map", 7, 5, false},
		{"Very large square map", 500, 500, false},
		{"Extremely large dimensions", 1000, 1000, false}, // Reduced from 10k to prevent test timeouts
		{"Single room map", roomMinSize + 2, roomMinSize + 2, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fg := NewFacilityGenerator(1234)
			m, px, py := fg.Generate(tc.width, tc.height)

			if tc.expectNilMap {
				if m != nil {
					t.Fatalf("Expected nil map, but got non-nil")
				}
				if px != 0 || py != 0 {
					t.Fatalf("Expected player position to be (0, 0), got (%d, %d)", px, py)
				}
				return
			}

			if m == nil {
				t.Fatalf("Expected non-nil map, but got nil")
			}

			if m.Width != tc.width || m.Height != tc.height {
				t.Fatalf("Expected map dimensions to be (%d, %d), got (%d, %d)", tc.width, tc.height, m.Width, m.Height)
			}

			// Ensure no overlapping rooms
			for i, room1 := range m.Rooms {
				for j, room2 := range m.Rooms {
					if i != j && room1.Intersects(room2) {
						t.Fatalf("Rooms %d and %d should not overlap", i, j)
					}
				}
			}

			// Verify that corridors connect all rooms
			for i := 1; i < len(m.Rooms); i++ {
				prevRoom := m.Rooms[i-1]
				currRoom := m.Rooms[i]

				prevCenterX, prevCenterY := prevRoom.Center()
				currCenterX, currCenterY := currRoom.Center()

				if !isCorridorConnected(m, prevCenterX, prevCenterY, currCenterX, currCenterY) {
					t.Fatalf("Room %d and Room %d are not connected by a corridor", i-1, i)
				}
			}

			countWalls, walkableTiles := countTiles(m)

			// Verify no tiles were lost or left unassigned
			if walkableTiles+countWalls != m.Width*m.Height {
				t.Fatalf("Tile count mismatch. Walkable (%d) + Walls (%d) != Total Area (%d)", walkableTiles, countWalls, m.Width*m.Height)
			}

			// Verify the player spawns in the first room
			if m.Rooms != nil && len(m.Rooms) > 0 {
				centerX, centerY := m.Rooms[0].Center()
				if px != centerX || py != centerY {
					t.Fatalf("Expected player position to match the first room center (%d, %d), got (%d, %d)", centerX, centerY, px, py)
				}
			}
		})
	}
}

func isCorridorConnected(m *Map, startX, startY, endX, endY int) bool {
	if m == nil {
		return false
	}

	start := m.GetTile(startX, startY)
	end := m.GetTile(endX, endY)
	if start == nil || end == nil {
		return false
	}
	if !start.Walkable || !end.Walkable {
		return false
	}

	type point struct{ x, y int }

	visited := make([]bool, m.Width*m.Height)
	idx := func(x, y int) int { return x + y*m.Width }

	q := make([]point, 0, 64)
	q = append(q, point{startX, startY})
	visited[idx(startX, startY)] = true

	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for head := 0; head < len(q); head++ {
		p := q[head]
		if p.x == endX && p.y == endY {
			return true
		}

		for _, d := range dirs {
			nx, ny := p.x+d[0], p.y+d[1]
			if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.Height {
				continue
			}
			i := idx(nx, ny)
			if visited[i] {
				continue
			}

			tile := m.GetTile(nx, ny)
			if tile == nil || !tile.Walkable {
				continue
			}

			visited[i] = true
			q = append(q, point{nx, ny})
		}
	}

	return false
}
