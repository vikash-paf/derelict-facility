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
	Visible  bool  // is the player looking at this tile right now?
	Explored bool  // has the player explored (ever visited) this tile?
	Variant  uint8 // Stores procedural noise/texture data
}
