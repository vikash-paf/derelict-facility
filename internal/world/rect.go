package world

type Rect struct {
	X1, Y1 int // left-top corner
	X2, Y2 int // right-bottom corner

	/*
		Left Edge: X1 (The smallest X value)
		Right Edge: X2 (The biggest X value)
		Top Edge: Y1 (The smallest Y value)
		Bottom Edge: Y2 (The biggest Y value)
	*/
}

func (r Rect) Center() (x, y int) {
	x = (r.X1 + r.X2) / 2
	y = (r.Y1 + r.Y2) / 2
	return x, y
}

func (r Rect) Width() int {
	return r.X2 - r.X1
}

func (r Rect) Height() int {
	return r.Y2 - r.Y1
}

func (r Rect) Intersects(other Rect) bool {
	// todo: implement AABB algo https://stackoverflow.com/a/22512393
	/*
		A's left edge is to the left of B's right edge.
		A's right edge is to the right of B's left edge.
		A's bottom edge is below B's top edge.
		A's top edge is above B's bottom edge.
	*/

	return false
}
