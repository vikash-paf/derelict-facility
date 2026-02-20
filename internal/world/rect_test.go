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
			overlaps: true,
		},
		{
			name:     "single point overlap",
			rect1:    Rect{X1: 5, Y1: 5, X2: 10, Y2: 10},
			rect2:    Rect{X1: 10, Y1: 10, X2: 15, Y2: 15},
			overlaps: true,
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

func TestRect_Center(t *testing.T) {
	tests := []struct {
		name  string
		r     Rect
		wantX int
		wantY int
	}{
		{
			name:  "odd span centered",
			r:     Rect{X1: 0, Y1: 0, X2: 9, Y2: 9},
			wantX: 4,
			wantY: 4,
		},
		{
			name:  "even span uses integer division",
			r:     Rect{X1: 0, Y1: 0, X2: 10, Y2: 10},
			wantX: 5,
			wantY: 5,
		},
		{
			name:  "negative coordinates",
			r:     Rect{X1: -10, Y1: -6, X2: -2, Y2: 4},
			wantX: -6,
			wantY: -1,
		},
		{
			name:  "mixed coordinates",
			r:     Rect{X1: -3, Y1: 2, X2: 4, Y2: 9},
			wantX: 0,
			wantY: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := tt.r.Center()
			if gotX != tt.wantX || gotY != tt.wantY {
				t.Fatalf("Center() = (%d,%d), want (%d,%d) for rect=%+v", gotX, gotY, tt.wantX, tt.wantY, tt.r)
			}
		})
	}
}
