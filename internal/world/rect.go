package world

type Rect struct {
	X1, Y1 int // top left corner
	X2, Y2 int // bottom right corner
}

func (r *Rect) Center() (x, y int) {
	x = (r.X1 + r.X2) / 2
	y = (r.Y1 + r.Y2) / 2
	return x, y
}

func (r *Rect) Width() int {
	return r.X2 - r.X1
}

func (r *Rect) Height() int {
	return r.Y2 - r.Y1
}

func (r *Rect) Intersects(other Rect) bool {
	// todo: check if the rectangles intersect
	return false
}
