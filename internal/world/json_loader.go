package world

import (
	"encoding/json"
	"os"

	"github.com/vikash-paf/derelict-facility/internal/entity"
)

// JSONMapData represents the structure of the map in a JSON file.
type JSONMapData struct {
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Rows   []string `json:"rows"`
}

// JSONMapLoader handles loading a world map from a JSON file.
type JSONMapLoader struct{}

func NewJSONMapLoader() *JSONMapLoader {
	return &JSONMapLoader{}
}

// Load reads the JSON file from OS disk and returns the Map, PlayerX, and PlayerY.
func (l *JSONMapLoader) Load(filepath string) (*Map, int, int, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, 0, 0, err
	}
	return l.LoadBytes(data)
}

// LoadBytes parses raw JSON map bytes and returns the Map, PlayerX, and PlayerY.
func (l *JSONMapLoader) LoadBytes(data []byte) (*Map, int, int, error) {

	var mapData JSONMapData
	if err := json.Unmarshal(data, &mapData); err != nil {
		return nil, 0, 0, err
	}

	m := NewMap(mapData.Width, mapData.Height)
	playerX, playerY := mapData.Width/2, mapData.Height/2
	var doors []entity.Point
	var terminals []entity.Point
	var generators []GeneratorInfo

	for y, row := range mapData.Rows {
		if y >= mapData.Height {
			break
		}
		for x, char := range row {
			if x >= mapData.Width {
				break
			}

			switch char {
			case '#': // Wall
				m.SetTile(x, y, Tile{Type: TileTypeWall, Walkable: false})
			case '.': // Floor
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
			case 'S', '*': // Skylight Glass Roof Floor
				yield := "O2_CAPSULE"
				if (x+y)%2 == 0 {
					yield = "MEDPACK"
				}
				m.SetTile(x, y, Tile{
					Type:              TileTypeFloor,
					Walkable:          true,
					IsSunlit:          true,
					Variant:           1,
					PlantStage:        PlantStageSeed,
					GrowthProgress:    0.0,
					GrowthRate:        1.5,
					YieldItemType:     yield,
				})

			case '+': // Door
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				doors = append(doors, entity.Point{X: x, Y: y})
			case 'T': // Terminal
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				terminals = append(terminals, entity.Point{X: x, Y: y})
			case 'G': // Global Power Generator
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				generators = append(generators, GeneratorInfo{Pos: entity.Point{X: x, Y: y}, IsGlobal: true})
			case 'g': // Local Room Power Generator
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				generators = append(generators, GeneratorInfo{Pos: entity.Point{X: x, Y: y}, IsGlobal: false})
			case '@': // Player start position
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				playerX, playerY = x, y
			case '>': // Down Stairway / Elevator to next level
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				m.Stairways = append(m.Stairways, StairwayInfo{
					Pos:  entity.Point{X: x, Y: y},
					IsUp: false,
				})
			case '<': // Up Stairway / Elevator to previous level
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				m.Stairways = append(m.Stairways, StairwayInfo{
					Pos:  entity.Point{X: x, Y: y},
					IsUp: true,
				})
			default:
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
			}
		}
	}

	m.Doors = doors
	m.Terminals = terminals
	m.PowerGenerators = generators

	// Automatically derive rooms from the floor layout
	l.deriveRooms(m)

	// Calculate bitmasks for walls so they tile properly
	l.calculateWallBitmasks(m)

	return m, playerX, playerY, nil
}

// deriveRooms uses a flood-fill algorithm to find contiguous floor areas.
func (l *JSONMapLoader) deriveRooms(m *Map) {
	visited := make([]bool, m.Width*m.Height)

	// Create a quick lookup for door locations to use as boundaries
	isDoor := make([]bool, m.Width*m.Height)
	for _, d := range m.Doors {
		isDoor[d.Y*m.Width+d.X] = true
	}

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			idx := y*m.Width + x
			tile := m.GetTile(x, y)

			// Start a fill if we find an unvisited floor tile that isn't a door
			if tile != nil && tile.Type == TileTypeFloor && !visited[idx] && !isDoor[idx] {
				minX, minY := x, y
				maxX, maxY := x, y

				// Simple BFS for flood fill
				queue := []entity.Point{{X: x, Y: y}}
				visited[idx] = true

				for len(queue) > 0 {
					p := queue[0]
					queue = queue[1:]

					if p.X < minX {
						minX = p.X
					}
					if p.Y < minY {
						minY = p.Y
					}
					if p.X > maxX {
						maxX = p.X
					}
					if p.Y > maxY {
						maxY = p.Y
					}

					// Check 4 neighbors
					neighbors := []entity.Point{
						{X: p.X, Y: p.Y - 1}, {X: p.X, Y: p.Y + 1},
						{X: p.X - 1, Y: p.Y}, {X: p.X + 1, Y: p.Y},
					}

					for _, n := range neighbors {
						if n.X < 0 || n.X >= m.Width || n.Y < 0 || n.Y >= m.Height {
							continue
						}
						nIdx := n.Y*m.Width + n.X
						nTile := m.GetTile(n.X, n.Y)
						if nTile != nil && nTile.Type == TileTypeFloor && !visited[nIdx] && !isDoor[nIdx] {
							visited[nIdx] = true
							queue = append(queue, n)
						}
					}
				}

				// If the area is big enough to be a room, add it
				if (maxX-minX) >= 1 && (maxY-minY) >= 1 {
					m.Rooms = append(m.Rooms, Rect{
						X1: minX, Y1: minY,
						X2: maxX, Y2: maxY,
					})
				}
			}
		}
	}
}

// calculateWallBitmasks sets the bitmask for walls to connect seamlessly.
func (l *JSONMapLoader) calculateWallBitmasks(m *Map) {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			tile := m.GetTile(x, y)
			if tile == nil || tile.Type != TileTypeWall {
				continue
			}

			var mask uint8 = 0

			// North
			if tN := m.GetTile(x, y-1); tN != nil && tN.Type == TileTypeWall {
				mask |= WallConnectionNorth
			}
			// East
			if tE := m.GetTile(x+1, y); tE != nil && tE.Type == TileTypeWall {
				mask |= WallConnectionEast
			}
			// South
			if tS := m.GetTile(x, y+1); tS != nil && tS.Type == TileTypeWall {
				mask |= WallConnectionSouth
			}
			// West
			if tW := m.GetTile(x-1, y); tW != nil && tW.Type == TileTypeWall {
				mask |= WallConnectionWest
			}

			tile.Bitmask = mask
		}
	}
}
