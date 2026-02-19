package world

// Tile represents a single square in the memory layout grid
type Tile struct {
	IsWall bool
}

// World manages the memory layout, grid, and tiles
type World struct {
	Width, Height int
	Grid          [][]Tile
}

// New initializes a new world map
func New(width, height int) *World {
	grid := make([][]Tile, height)
	for y := range grid {
		grid[y] = make([]Tile, width)
	}

	return &World{
		Width:  width,
		Height: height,
		Grid:   grid,
	}
}
