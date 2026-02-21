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

func (m *Map) IsWalkable(x, y int) bool {
	tile := m.GetTile(x, y)
	if tile == nil {
		return false
	}
	return tile.Walkable
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
	for i := range m.Tiles {
		m.Tiles[i].Visible = false
	}

	// clamp the bounding box so we stay inside the map
	minX := max(0, playerX-radius)
	maxX := min(m.Width-1, playerX+radius)
	minY := max(0, playerY-radius)
	maxY := min(m.Height-1, playerY+radius)

	// cast rays only to the parts of the perimeter that exist
	for x := minX; x <= maxX; x++ {
		m.castRay(playerX, playerY, x, minY) // Top edge
		m.castRay(playerX, playerY, x, maxY) // Bottom edge
	}

	for y := minY; y <= maxY; y++ {
		m.castRay(playerX, playerY, minX, y) // Left edge
		m.castRay(playerX, playerY, maxX, y) // Right edge
	}

	// the player can always see their own tile
	m.Tiles[playerX+playerY*m.Width].Visible = true
}

func (m *Map) castRay(x1, y1, x2, y2 int) {
	// implement cast the "ray" using Bresenham's line algorithm'

	line := getLine(x1, y1, x2, y2)

	for _, p := range line {
		// if the ray goes off the map, stop
		if p.X < 0 || p.X >= m.Width || p.Y < 0 || p.Y >= m.Height {
			break
		}

		idx := p.X + p.Y*m.Width

		m.Tiles[idx].Visible = true
		m.Tiles[idx].Explored = true

		// if the tile is not walkable then light can't pass through
		if !m.Tiles[idx].Walkable {
			break
		}
	}
}
