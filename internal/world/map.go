package world

type Map struct {
	Tiles  []Tile
	Rooms  []Rect
	Width  int
	Height int
}

func NewMap(width, height int) *Map {
	return &Map{
		Width:  width,
		Height: height,
		Tiles:  make([]Tile, width*height),
	}
}

func (m *Map) SetTile(x, y int, tile Tile) {
	m.Tiles[x+y*m.Width] = tile
}

func (m *Map) GetTile(x, y int) *Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return nil
	}

	// basic math to calculate the index of a tile in the "tape"
	return &m.Tiles[x+y*m.Width]
}

func (m *Map) ComputeFOV(playerX, playerY int, radius int) {
	// 1. reset all tiles to not visible
	// 2. compute the visible tiles
	// 3. mark all tiles in the radius as visible
	// uses Bresenham's line algorithm
}
