package world

import (
	"math/rand/v2"
)

const (
	maxRooms    = 30
	roomMinSize = 4
	roomMaxSize = 10
)

func NewFacilityGenerator(seed uint64) *FacilityGenerator {
	return &FacilityGenerator{
		seed: seed,
		rng:  rand.New(rand.NewPCG(seed, seed)),
	}
}

type FacilityGenerator struct {
	seed uint64
	rng  *rand.Rand // random number generator
}

func (f FacilityGenerator) Generate(width, height int) (*Map, int, int) {
	if width < roomMinSize || height < roomMinSize {
		return nil, 0, 0
	}

	// 1. create the empty map, with walls (non walkable tiles)
	m := NewMap(width, height)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			m.SetTile(x, y, Tile{Type: TileTypeWall, Walkable: false})
		}
	}

	// 2. run the generation algorithm (l-Corridors algorithm, aka Procedural Dungeon Generator)
	// 2.1 carve the rooms
	// 2.2 connect the rooms (l-corridors)
	// 3. return the map and the player position (center of the first room)
	return m, 0, 0
}
