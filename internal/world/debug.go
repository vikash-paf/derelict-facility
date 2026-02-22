package world

import (
	"fmt"
	"strings"

	"github.com/vikash-paf/derelict-facility/internal/entity"
)

// InspectVisibility Helper to render the visibility state to the console
func InspectVisibility(m *Map, px int, py int) string {
	var sb strings.Builder
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if x == px && y == py {
				sb.WriteString("P ") // Player position
				continue
			}
			tile := m.Tiles[x+y*m.Width]
			if tile.Visible {
				if !tile.Walkable {
					sb.WriteString("X ") // Wall that light hit
				} else {
					sb.WriteString("V ") // Visible floor
				}
			} else {
				if !tile.Walkable {
					sb.WriteString("# ") // Unseen wall
				} else {
					sb.WriteString(". ") // Unseen floor
				}
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// VisualizePath prints the map state to the console for debugging.
func VisualizePath(m *Map, path []entity.Point, start, target entity.Point) {
	// Create a temporary overlay for the path
	pathMap := make(map[entity.Point]bool)
	for _, p := range path {
		pathMap[p] = true
	}

	fmt.Printf("\n--- Path Visualization (%dx%d) ---\n", m.Width, m.Height)
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			p := entity.Point{X: x, Y: y}
			tile := m.GetTile(x, y)
			if tile == nil {
				continue
			}

			char := " . " // Default Floor
			if !tile.Walkable {
				char = "###" // Wall
			}
			if pathMap[p] {
				char = " o " // Path
			}
			if p == start {
				char = " S " // Start
			}
			if p == target {
				char = " G " // Goal
			}

			fmt.Print(char)
		}
		fmt.Println()
	}

	fmt.Println("---------------------------------")
}
