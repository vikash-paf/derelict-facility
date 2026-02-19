package world

type TileType int

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
	Tiles  []Tile
}
