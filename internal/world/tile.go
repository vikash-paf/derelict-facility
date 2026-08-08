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
	IsSunlit          bool    // true if tile is situated under a glass skylight roof corridor
	SunlightIntensity float64 // 1.0 under skylight roof, decays outwards (0.5, 0.25) through open doors/corridors
	Variant           uint8   // Stores procedural noise/texture data
	Distance int   // Distance from the player when last visible (for depth shading)
	Bitmask  uint8 // 4-bit mask storing neighbor data for auto-tiling walls
}
