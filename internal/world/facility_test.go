package world

import (
	"testing"
)

func TestFacilityGenerator_Generate(t *testing.T) {
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
			name:          "Small square map",
			width:         4,
			height:        4,
			expectedWalls: 16,
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
			name:          "Large square map",
			width:         50,
			height:        50,
			expectedWalls: 2500,
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

			countWalls := 0
			for x := 0; x < m.Width; x++ {
				for y := 0; y < m.Height; y++ {
					tile := m.GetTile(x, y)
					if tile.Type == TileTypeWall && !tile.Walkable {
						countWalls++
					}
				}
			}

			if countWalls != tc.expectedWalls {
				t.Fatalf("Expected %d wall tiles, but counted %d", tc.expectedWalls, countWalls)
			}

			if px != 0 || py != 0 {
				t.Fatalf("Expected player position to be (0, 0), got (%d, %d)", px, py)
			}
		})
	}
}
