package world

import (
	"github.com/vikash-paf/derelict-facility/internal/entity"
)

type GeneratorInfo struct {
	Pos      entity.Point
	IsGlobal bool
}

type StairwayInfo struct {
	Pos               entity.Point
	TargetLevelID     string
	RequiredClearance uint32
	IsUp              bool
}

type Map struct {
	Tiles           []Tile
	Rooms           []Rect
	Doors           []entity.Point
	Terminals       []entity.Point
	PowerGenerators []GeneratorInfo
	Stairways       []StairwayInfo
	Width           int
	Height          int
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

// GetIndex converts 2D coordinates into a 1D array index.
func (m *Map) GetIndex(x, y int) int {
	return y*m.Width + x
}

// GetIndexFromPoint is a convenience wrapper for our Point struct.
func (m *Map) GetIndexFromPoint(p entity.Point) int {
	return p.Y*m.Width + p.X
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

func (m *Map) ComputeFOV(playerX, playerY int, radius int, blocksLight func(x, y int) bool, isTilePowered func(x, y int) bool) {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			idx := m.GetIndex(x, y)
			if isTilePowered != nil && isTilePowered(x, y) {
				m.Tiles[idx].Visible = true
				m.Tiles[idx].Explored = true
				m.Tiles[idx].Distance = 0
			} else {
				m.Tiles[idx].Visible = false
			}
		}
	}

	// clamp the bounding box so we stay inside the map
	minX := max(0, playerX-radius)
	maxX := min(m.Width-1, playerX+radius)
	minY := max(0, playerY-radius)
	maxY := min(m.Height-1, playerY+radius)

	// cast rays only to the parts of the perimeter that exist
	for x := minX; x <= maxX; x++ {
		m.castRay(playerX, playerY, x, minY, blocksLight) // Top edge
		m.castRay(playerX, playerY, x, maxY, blocksLight) // Bottom edge
	}

	for y := minY; y <= maxY; y++ {
		m.castRay(playerX, playerY, minX, y, blocksLight) // Left edge
		m.castRay(playerX, playerY, maxX, y, blocksLight) // Right edge
	}

	// the player can always see their own tile
	m.Tiles[playerX+playerY*m.Width].Visible = true
	m.Tiles[playerX+playerY*m.Width].Distance = 0
}

func (m *Map) castRay(x1, y1, x2, y2 int, blocksLight func(x, y int) bool) {
	// implement cast the "ray" using Bresenham's line algorithm'

	getLine(x1, y1, x2, y2, func(x, y int) bool {
		// if the ray goes off the map, stop
		if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
			return false
		}

		idx := x + y*m.Width

		m.Tiles[idx].Visible = true
		m.Tiles[idx].Explored = true

		// Approximate distance using chebyshev/manhattan or simple max component
		dx := x - x1
		if dx < 0 {
			dx = -dx
		}
		dy := y - y1
		if dy < 0 {
			dy = -dy
		}
		dist := dx
		if dy > dx {
			dist = dy
		}
		m.Tiles[idx].Distance = dist

		// Use the callback to decide if light passes through
		if blocksLight(x, y) {
			return false
		}

		return true // continue drawing the line
	})
}

// PropagateSunlight uses BFS to spill sunlight from glass roof tiles into nearby open floor tiles.
func (m *Map) PropagateSunlight(maxSpillRadius int, isSolid func(x, y int) bool) {
	type lightNode struct {
		x, y  int
		dist  int
		intensity float64
	}

	queue := []lightNode{}
	visited := make(map[int]float64)

	// Seed BFS with all glass roof tiles (IsSunlit == true)
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			idx := m.GetIndex(x, y)
			tile := &m.Tiles[idx]
			if tile.IsSunlit {
				tile.SunlightIntensity = 1.0
				queue = append(queue, lightNode{x: x, y: y, dist: 0, intensity: 1.0})
				visited[idx] = 1.0
			} else {
				tile.SunlightIntensity = 0.0
			}
		}
	}

	// BFS outward spillover through unblocked tiles
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.dist >= maxSpillRadius {
			continue
		}

		neighbors := [4][2]int{
			{curr.x + 1, curr.y}, {curr.x - 1, curr.y},
			{curr.x, curr.y + 1}, {curr.x, curr.y - 1},
		}

		nextIntensity := curr.intensity * 0.5 // Decays by 50% per step

		for _, n := range neighbors {
			nx, ny := n[0], n[1]
			if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.Height {
				continue
			}

			// Sunlight cannot spill through solid walls or closed doors
			if !m.IsWalkable(nx, ny) || (isSolid != nil && isSolid(nx, ny)) {
				continue
			}

			nIdx := m.GetIndex(nx, ny)
			if prevIntensity, exists := visited[nIdx]; !exists || nextIntensity > prevIntensity {
				visited[nIdx] = nextIntensity
				m.Tiles[nIdx].SunlightIntensity = nextIntensity
				queue = append(queue, lightNode{x: nx, y: ny, dist: curr.dist + 1, intensity: nextIntensity})
			}
		}
	}
}
