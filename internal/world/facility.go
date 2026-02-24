package world

import (
	"math/rand/v2"

	"github.com/vikash-paf/derelict-facility/internal/entity"
)

const (
	maxRooms    = 50
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
	rng  *rand.Rand
	seed uint64
}

func (f *FacilityGenerator) randomBetween(min, max int) int {
	spread := max - min + 1
	return f.rng.IntN(spread) + min
}

func (f *FacilityGenerator) carveFloor(m *Map, x, y int) {
	// 80% chance for a blank space, 20% chance for a random noise texture
	variant := uint8(0)
	if f.rng.IntN(10) < 2 {
		variant = uint8(f.rng.IntN(4) + 1) // Variants 1 to 4
	}
	m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: variant})
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

		// If the max is less than 0, the newRoom is too big to fit.
		// Skip this attempt and try rolling a new newRoom!
		if maxX < 0 || maxY < 0 {
			continue
		}

		x := f.randomBetween(0, maxX)
		y := f.randomBetween(0, maxY)

		newRoom := Rect{X1: x, Y1: y, X2: x + rWidth, Y2: y + rHeight}

		// carve newRoom into the map
		overlaps := false
		for _, otherRoom := range rooms {
			if newRoom.Intersects(otherRoom) {
				overlaps = true
				break
			}
		}

		if overlaps {
			continue
		}
		// carve newRoom
		for rx := newRoom.X1; rx <= newRoom.X2; rx++ {
			for ry := newRoom.Y1; ry <= newRoom.Y2; ry++ {
				f.carveFloor(m, rx, ry)
			}
		}

		// get player coordinates
		if len(rooms) == 0 {
			playerX, playerY = newRoom.Center()
		} else {
			prevRoom := rooms[len(rooms)-1]
			prevX, prevY := prevRoom.Center()
			newX, newY := newRoom.Center()

			if f.rng.IntN(2) == 1 {
				// Horizontal then Vertical
				f.createHorizontalCorridor(m, prevX, newX, prevY)
				f.createVerticalCorridor(m, prevY, newY, newX)
			} else {
				// Vertical then Horizontal
				f.createVerticalCorridor(m, prevY, newY, prevX)
				f.createHorizontalCorridor(m, prevX, newX, newY)
			}
		}

		rooms = append(rooms, newRoom)
	}

	// 2. run the generation algorithm (l-Corridors algorithm, aka Procedural Dungeon Generator)
	// 2.1 carve the rooms
	// 2.2 connect the rooms (l-corridors)
	// 3. run auto-tiling calculation for all walls
	f.calculateWallBitmasks(m)

	m.Rooms = rooms
	m.Doors = f.findDoorways(m)

	return m, playerX, playerY
}

func (f FacilityGenerator) findDoorways(m *Map) []entity.Point {
	var doors []entity.Point
	seen := make(map[entity.Point]bool)

	for _, room := range m.Rooms {
		// A room's perimeter is one tile outside its boundary
		// Check top and bottom edges
		for x := room.X1; x <= room.X2; x++ {
			p1 := entity.Point{X: x, Y: room.Y1 - 1}
			if m.IsWalkable(p1.X, p1.Y) && !seen[p1] {
				seen[p1] = true
				doors = append(doors, p1)
			}
			p2 := entity.Point{X: x, Y: room.Y2 + 1}
			if m.IsWalkable(p2.X, p2.Y) && !seen[p2] {
				seen[p2] = true
				doors = append(doors, p2)
			}
		}
		// Check left and right edges
		for y := room.Y1; y <= room.Y2; y++ {
			p1 := entity.Point{X: room.X1 - 1, Y: y}
			if m.IsWalkable(p1.X, p1.Y) && !seen[p1] {
				seen[p1] = true
				doors = append(doors, p1)
			}
			p2 := entity.Point{X: room.X2 + 1, Y: y}
			if m.IsWalkable(p2.X, p2.Y) && !seen[p2] {
				seen[p2] = true
				doors = append(doors, p2)
			}
		}
	}
	return doors
}

func (f FacilityGenerator) calculateWallBitmasks(m *Map) {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			tile := m.GetTile(x, y)
			if tile == nil || tile.Type != TileTypeWall {
				continue
			}

			var mask uint8 = 0

			// North
			tN := m.GetTile(x, y-1)
			if tN != nil && tN.Type == TileTypeWall {
				mask |= 1 // 0001
			}
			// East
			tE := m.GetTile(x+1, y)
			if tE != nil && tE.Type == TileTypeWall {
				mask |= 2 // 0010
			}
			// South
			tS := m.GetTile(x, y+1)
			if tS != nil && tS.Type == TileTypeWall {
				mask |= 4 // 0100
			}
			// West
			tW := m.GetTile(x-1, y)
			if tW != nil && tW.Type == TileTypeWall {
				mask |= 8 // 1000
			}

			tile.Bitmask = mask
		}
	}
}

func (f FacilityGenerator) createHorizontalCorridor(m *Map, x1, x2, y int) {
	// Swap if drawing right-to-left
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		f.carveFloor(m, x, y)
	}
}

func (f FacilityGenerator) createVerticalCorridor(m *Map, y1, y2, x int) {
	// Swap if drawing bottom-to-top
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		f.carveFloor(m, x, y)
	}
}
