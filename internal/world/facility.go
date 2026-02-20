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

func (f *FacilityGenerator) randomBetween(min, max int) int {
	spread := max - min + 1
	return f.rng.IntN(spread) + min
}

func (f FacilityGenerator) Generate(width, height int) (*Map, int, int) {
	if width < roomMinSize || height < roomMinSize {
		return nil, 0, 0
	}

	playerX, playerY := width/2, height/2

	// 1. create the empty map, with walls (non walkable tiles)
	m := NewMap(width, height)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			m.SetTile(x, y, Tile{Type: TileTypeWall, Walkable: false})
		}
	}

	var rooms []Rect

	for i := 0; i < maxRooms; i++ {
		rWidth := f.randomBetween(roomMinSize, roomMaxSize)
		rHeight := f.randomBetween(roomMinSize, roomMaxSize)

		maxX := width - rWidth - 1
		maxY := height - rHeight - 1

		// If the max is less than 0, the room is too big to fit.
		// Skip this attempt and try rolling a new room!
		if maxX < 0 || maxY < 0 {
			continue
		}

		x := f.randomBetween(0, maxX)
		y := f.randomBetween(0, maxY)

		room := Rect{X1: x, Y1: y, X2: x + rWidth, Y2: y + rHeight}

		// carve room into the map
		for r := range rooms {
			if room.Intersects(rooms[r]) {
				continue
			} else {
				// carve room
				for x := room.X1; x <= room.X2; x++ {
					for y := room.Y1; y <= room.Y2; y++ {
						m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true})
					}
				}
				rooms = append(rooms, room)

				// get player coordinates
				if len(rooms) == 1 {
					playerX, playerY = room.Center()
				}
			}
		}
		// If intersects throw it away
		// continue
		// else carve it
	}

	// 2. run the generation algorithm (l-Corridors algorithm, aka Procedural Dungeon Generator)
	// 2.1 carve the rooms
	// 2.2 connect the rooms (l-corridors)
	// 3. return the map and the player position (center of the first room)
	return m, playerX, playerY
}
