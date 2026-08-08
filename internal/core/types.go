package core

type InputEvent struct {
	Key  rune
	Code int // for non-runes
	Quit bool
}

// Camera defines a viewport into the world.
// X and Y represent the top-left corner of the camera in WORLD coordinates.
// Width and Height are the dimensions of the viewport in SCREEN coordinates.
type Camera struct {
	X      int
	Y      int
	Width  int
	Height int
}

// WorldToScreen takes a coordinate from the game map and converts it
// to a coordinate on your monitor, relative to the camera's position.
func (c *Camera) WorldToScreen(worldX, worldY int) (screenX, screenY int) {
	screenX = worldX - c.X
	screenY = worldY - c.Y
	return screenX, screenY
}

// IsInView returns true if the world coordinate is currently visible inside the camera's viewport.
func (c *Camera) IsInView(worldX, worldY int) bool {
	screenX, screenY := c.WorldToScreen(worldX, worldY)
	return screenX >= 0 && screenX < c.Width && screenY >= 0 && screenY < c.Height
}
