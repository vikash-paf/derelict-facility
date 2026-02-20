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

	calculateExpectedWalls := func(width, height int, rooms []Rect) int {
		walls := width * height
		for _, room := range rooms {
			walls -= room.Width() * room.Height()
		}
		return walls
	}
	tests := []struct {
		name          string
		width, height int
		expectedWalls int
		expectNilMap  bool
	}{
		{
			name:          "Zero dimensions",
			width:         0,
			height:        0,
			expectedWalls: 0,
			expectNilMap:  true,
		},
		{
			name:          "Minimal valid map",
			width:         roomMinSize,
			height:        roomMinSize,
			expectedWalls: roomMinSize * roomMinSize,
			expectNilMap:  false,
		},
		{
			name:          "Small square map",
			width:         4,
			height:        4,
			expectedWalls: 16,
			expectNilMap:  false,
		},
		{
			name:          "Wide map with minimal height",
			width:         20,
			height:        roomMinSize,
			expectedWalls: 20 * roomMinSize,
			expectNilMap:  false,
		},
		{
			name:          "Tall map with minimal width",
			width:         roomMinSize,
			height:        30,
			expectedWalls: roomMinSize * 30,
			expectNilMap:  false,
		},
		{
			name:          "Rectangular map",
			width:         7,
			height:        5,
			expectedWalls: 35,
			expectNilMap:  false,
		},
		{
			name:          "Very large square map",
			width:         500,
			height:        500,
			expectedWalls: 500 * 500,
			expectNilMap:  false,
		},
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

			countWalls, walkableTiles := countTiles(m)
			if walkableTiles != (m.Width*m.Height - countWalls) {
				t.Fatalf("Walkable tile count (%d) does not match expected value (%d)", walkableTiles, m.Width*m.Height-countWalls)
			}

			expectedWalls := calculateExpectedWalls(tc.width, tc.height, m.Rooms)
			if countWalls != expectedWalls {
				t.Fatalf("Expected %d wall tiles after room carving, but counted %d", expectedWalls, countWalls)
			}

			if m.Rooms != nil && len(m.Rooms) > 0 {
				centerX, centerY := m.Rooms[0].Center()
				if px != centerX || py != centerY {
					t.Fatalf("Expected player position to match the first room center (%d, %d), got (%d, %d)", centerX, centerY, px, py)
				}
			}
		})
	}
}
