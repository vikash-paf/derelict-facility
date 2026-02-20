package world

import "testing"

func TestRect_Intersects(t *testing.T) {
	tests := []struct {
		name     string
		rect1    Rect
		rect2    Rect
		overlaps bool
	}{
		{
			name:     "fully overlapping",
			rect1:    Rect{X1: 0, Y1: 0, X2: 10, Y2: 10},
			rect2:    Rect{X1: 2, Y1: 2, X2: 8, Y2: 8},
			overlaps: true,
		},
		{
			name:     "edge touching",
			rect1:    Rect{X1: 0, Y1: 0, X2: 10, Y2: 10},
			rect2:    Rect{X1: 10, Y1: 0, X2: 20, Y2: 10},
			overlaps: true,
		},
		{
			name:     "no overlap",
			rect1:    Rect{X1: 0, Y1: 0, X2: 5, Y2: 5},
			rect2:    Rect{X1: 10, Y1: 10, X2: 15, Y2: 15},
			overlaps: false,
		},
		{
			name:     "intersecting edges",
			rect1:    Rect{X1: 0, Y1: 0, X2: 10, Y2: 10},
			rect2:    Rect{X1: 5, Y1: 5, X2: 15, Y2: 15},
			overlaps: true,
		},
		{
			name:     "completely outside",
			rect1:    Rect{X1: 0, Y1: 0, X2: 5, Y2: 5},
			rect2:    Rect{X1: 6, Y1: 6, X2: 10, Y2: 10},
			overlaps: false,
		},
		{
			name:     "shared corner",
			rect1:    Rect{X1: 0, Y1: 0, X2: 10, Y2: 10},
			rect2:    Rect{X1: 10, Y1: 10, X2: 15, Y2: 15},
			overlaps: false,
		},
		{
			name:     "single point overlap",
			rect1:    Rect{X1: 5, Y1: 5, X2: 10, Y2: 10},
			rect2:    Rect{X1: 10, Y1: 10, X2: 15, Y2: 15},
			overlaps: false,
		},
		{
			name:     "negative coordinates overlap",
			rect1:    Rect{X1: -10, Y1: -10, X2: 0, Y2: 0},
			rect2:    Rect{X1: -5, Y1: -5, X2: 5, Y2: 5},
			overlaps: true,
		},
		{
			name:     "negative coordinates no overlap",
			rect1:    Rect{X1: -10, Y1: -10, X2: -5, Y2: -5},
			rect2:    Rect{X1: 0, Y1: 0, X2: 5, Y2: 5},
			overlaps: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.rect1.Intersects(tt.rect2)

			if actual != tt.overlaps {
				t.Errorf("Rect1: %+v, Rect2: %+v, expected %v, got %v", tt.rect1, tt.rect2, tt.overlaps, actual)
			}
		})
	}
}
