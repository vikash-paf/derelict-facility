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

// Load reads the JSON file and returns the Map, PlayerX, and PlayerY.
func (l *JSONMapLoader) Load(filepath string) (*Map, int, int, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, 0, 0, err
	}

	var mapData JSONMapData
	if err := json.Unmarshal(data, &mapData); err != nil {
		return nil, 0, 0, err
	}

	m := NewMap(mapData.Width, mapData.Height)
	playerX, playerY := mapData.Width/2, mapData.Height/2
	var doors []entity.Point
	var terminals []entity.Point
	var generators []entity.Point

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
			case 'D': // Door
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				doors = append(doors, entity.Point{X: x, Y: y})
			case 'T': // Terminal
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				terminals = append(terminals, entity.Point{X: x, Y: y})
			case 'G': // Power Generator
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				generators = append(generators, entity.Point{X: x, Y: y})
			case '@': // Player start position
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
				playerX, playerY = x, y
			default:
				m.SetTile(x, y, Tile{Type: TileTypeFloor, Walkable: true, Variant: 1})
			}
		}
	}

	m.Doors = doors
	m.Terminals = terminals
	m.PowerGenerators = generators

	// Calculate bitmasks for walls so they tile properly
	l.calculateWallBitmasks(m)

	return m, playerX, playerY, nil
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
				mask |= 1
			}
			// East
			if tE := m.GetTile(x+1, y); tE != nil && tE.Type == TileTypeWall {
				mask |= 2
			}
			// South
			if tS := m.GetTile(x, y+1); tS != nil && tS.Type == TileTypeWall {
				mask |= 4
			}
			// West
			if tW := m.GetTile(x-1, y); tW != nil && tW.Type == TileTypeWall {
				mask |= 8
			}

			tile.Bitmask = mask
		}
	}
}
