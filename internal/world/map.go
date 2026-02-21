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

	for i := range m.Tiles {
		m.Tiles[i].Visible = false
	}

	// 1.1 get the (light radius) bounding box
	minX, maxX := playerX-radius, playerX+radius
	minY, maxY := playerY-radius, playerY+radius

	// 1.2 cast rays to the points in the bounding box (TOP and BOTTOM edges of the box)
	for x := minX; x <= maxX; x++ {
		// cast ray to top edge
		m.castRay(playerX, playerY, x, minY)
		// cast ray to bottom edge
		m.castRay(playerX, playerY, x, maxY)

	}

	// 1.2 cast rays to the points in the bounding box (TOP and BOTTOM edges of the box)
	for y := minY; y <= maxY; y++ {
		// cast ray to top edge
		m.castRay(playerX, playerY, minX, y)
		// cast ray to bottom edge
		m.castRay(playerX, playerY, maxX, y)
	}
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

func getLine(x1, y1, x2, y2 int) []Point {
	points := make([]Point, 0)

	// get all the coordinates of the "cells" in the line

	return points
}
