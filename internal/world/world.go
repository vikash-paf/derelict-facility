package world

type TileType uint8

const (
	TileTypeEmpty TileType = iota
	TileTypePlayer
	TileTypeWall
	TileTypeFloor
)

type Tile struct {
	Type     TileType
	Walkable bool
}

type Map struct {
	Width  int
	Height int

	// keeping this as a single slice for now,
	// this will store the location, it's continuous "tape", better performance and low memory usage
	Tiles []Tile
}
