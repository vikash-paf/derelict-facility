package world

import (
	"strings"
)

// InspectVisibility Helper to render the visibility state to the console
func InspectVisibility(m *Map) string {
	var sb strings.Builder
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
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
