package world

type TileType uint8

const (
	TileTypeEmpty TileType = iota
	TileTypeWall
	TileTypeFloor
)

type Tile struct {
	Type     TileType
	Walkable bool
}
