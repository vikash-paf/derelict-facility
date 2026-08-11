package world

type TileType uint8

const (
	TileTypeEmpty TileType = iota
	TileTypeWall
	TileTypeFloor
)

type PlantStage uint8

const (
	PlantStageNone PlantStage = iota
	PlantStageSeed
	PlantStageSprout
	PlantStageMature
)

type Tile struct {
	Type              TileType
	Walkable          bool
	Visible           bool    // is the player looking at this tile right now?
	Explored          bool    // has the player explored (ever visited) this tile?
	IsSunlit          bool    // true if tile is situated under a glass skylight roof corridor
	SunlightIntensity float64 // 1.0 under skylight roof, decays outwards (0.5, 0.25) through open doors/corridors
	Variant           uint8   // Stores procedural noise/texture data
	Distance          int     // Distance from the player when last visible (for depth shading)
	Bitmask           uint8   // 4-bit mask storing neighbor data for auto-tiling walls

	// Plant attributes
	PlantStage     PlantStage
	GrowthProgress float64
	GrowthRate     float64
	YieldItemType  string
}


// Cardinal Wall Connection Bitmasks
const (
	WallConnectionNone  uint8 = 0
	WallConnectionNorth uint8 = 1
	WallConnectionEast  uint8 = 2
	WallConnectionSouth uint8 = 4
	WallConnectionWest  uint8 = 8
)

// Combined Wall Auto-Tiling Bitmasks
const (
	WallBitmaskIsolated       uint8 = 0
	WallBitmaskNorthOnly      uint8 = 1
	WallBitmaskEastOnly       uint8 = 2
	WallBitmaskSouthOnly      uint8 = 4
	WallBitmaskWestOnly       uint8 = 8
	WallBitmaskNorthSouth     uint8 = 5
	WallBitmaskEastWest       uint8 = 10
	WallBitmaskNorthEast      uint8 = 3
	WallBitmaskEastSouth      uint8 = 6
	WallBitmaskSouthWest      uint8 = 12
	WallBitmaskWestNorth      uint8 = 9
	WallBitmaskNorthEastSouth uint8 = 7
	WallBitmaskEastSouthWest  uint8 = 14
	WallBitmaskSouthWestNorth uint8 = 13
	WallBitmaskWestNorthEast  uint8 = 11
	WallBitmaskIntersection   uint8 = 15
)
