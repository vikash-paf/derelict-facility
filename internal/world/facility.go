package world

import (
	"math/rand/v2"
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
	// todo:
	// 1. create the empty map
	// 2. run the generation algorithm (l-Corridors algorithm, aka Procedural Dungeon Generator)
	// 3. return the map and the player position (center of the first room)
	return nil, 0, 0
}
