package world

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetLine(t *testing.T) {
	tests := []struct {
		name     string
		x0, y0   int
		x1, y1   int
		expected []Point
	}{
		{
			name:     "horizontal-right",
			x0:       0,
			y0:       0,
			x1:       5,
			y1:       0,
			expected: []Point{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {5, 0}},
		},
		{
			name:     "horizontal-left",
			x0:       5,
			y0:       0,
			x1:       0,
			y1:       0,
			expected: []Point{{5, 0}, {4, 0}, {3, 0}, {2, 0}, {1, 0}, {0, 0}},
		},
		{
			name:     "vertical-up",
			x0:       0,
			y0:       0,
			x1:       0,
			y1:       5,
			expected: []Point{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}},
		},
		{
			name:     "vertical-down",
			x0:       0,
			y0:       5,
			x1:       0,
			y1:       0,
			expected: []Point{{0, 5}, {0, 4}, {0, 3}, {0, 2}, {0, 1}, {0, 0}},
		},
		{
			name:     "diagonal-right-up",
			x0:       0,
			y0:       0,
			x1:       5,
			y1:       5,
			expected: []Point{{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},
		},
		{
			name:     "diagonal-left-down",
			x0:       5,
			y0:       5,
			x1:       0,
			y1:       0,
			expected: []Point{{5, 5}, {4, 4}, {3, 3}, {2, 2}, {1, 1}, {0, 0}},
		},
		{
			name:     "diagonal-right-down",
			x0:       0,
			y0:       5,
			x1:       5,
			y1:       0,
			expected: []Point{{0, 5}, {1, 4}, {2, 3}, {3, 2}, {4, 1}, {5, 0}},
		},
		{
			name:     "diagonal-left-up",
			x0:       5,
			y0:       0,
			x1:       0,
			y1:       5,
			expected: []Point{{5, 0}, {4, 1}, {3, 2}, {2, 3}, {1, 4}, {0, 5}},
		},
		{
			name:     "single-point",
			x0:       3,
			y0:       3,
			x1:       3,
			y1:       3,
			expected: []Point{{3, 3}},
		},
		{
			name:     "steep-slope",
			x0:       2,
			y0:       3,
			x1:       4,
			y1:       8,
			expected: []Point{{2, 3}, {2, 4}, {3, 5}, {3, 6}, {4, 7}, {4, 8}},
		},
		{
			name:     "shallow-slope",
			x0:       3,
			y0:       2,
			x1:       8,
			y1:       4,
			expected: []Point{{3, 2}, {4, 2}, {5, 3}, {6, 3}, {7, 4}, {8, 4}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLine(tt.x0, tt.y0, tt.x1, tt.y1)

			if !reflect.DeepEqual(result, tt.expected) {
				fmt.Println(renderDebug(result, tt.x1-tt.x0, tt.y1-tt.y0))

				t.Errorf("getLine(%d, %d, %d, %d) = %v, want %v", tt.x0, tt.y0, tt.x1, tt.y1, result, tt.expected)
			}
		})
	}
}

func renderDebug(points []Point, w, h int) string {
	// Create a 2D grid of dots
	grid := make([][]string, h+1)
	for i := range grid {
		grid[i] = make([]string, w+1)
		for j := range grid[i] {
			grid[i][j] = "."
		}
	}
	// Plot the points with hashes
	for _, p := range points {
		if p.Y >= 0 && p.Y <= h && p.X >= 0 && p.X <= w {
			grid[p.Y][p.X] = "#"
		}
	}
	// Stringify
	var out string
	for y := h; y >= 0; y-- { // Flip Y so 0 is bottom for "natural" math
		for x := 0; x <= w; x++ {
			out += grid[y][x] + " "
		}
		out += "\n"
	}
	return out
}
