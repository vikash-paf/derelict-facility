package world

import (
	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/math"
)

// getLine uses Bresenham's Line Algorithm to return all points between A and B.
// Think of this as charting a laser beam or line-of-sight across your facility's floor grid.
// Read more: https://en.wikipedia.org/wiki/Bresenham%27s_line_algorithm
func getLine(x0, y0, x1, y1 int) []entity.Point {
	var points []entity.Point

	// Calculate the absolute distances between the start and end points.
	// dx is the total horizontal distance.
	dx := math.Abs(x1 - x0)

	// dy is the total vertical distance. We make it negative here as a math trick
	// so we can simply add it to our 'error' tracker later instead of subtracting.
	dy := -math.Abs(y1 - y0)

	// Determine the step direction for the X-axis (left or right).
	// Default to stepping left (-1). If the destination is to the right, change to stepping right (1).
	sx := -1
	if x0 < x1 {
		sx = 1
	}

	// Determine the step direction for the Y-axis (up or down).
	// Default to stepping up (-1). If the destination is below, change to stepping down (1).
	sy := -1
	if y0 < y1 {
		sy = 1
	}

	// Initialize the error tracker. This balances the scale, telling us how far
	// our current grid step has drifted from the true mathematical straight line.
	err := dx + dy

	// Enter the infinite loop to walk the grid, tile by tile.
	for {
		// Log the current coordinate. The terminal needs to know every single tile to light up.
		points = append(points, entity.Point{X: x0, Y: y0})

		// Check if we have reached the destination coordinates. If yes, terminate the loop.
		if x0 == x1 && y0 == y1 {
			break
		}

		// Store double the current error for comparison.
		// Doubling it prevents us from ever needing to use floating-point fractions (decimals).
		e2 := 2 * err

		// Adjust X if our drift tolerance allows it.
		if e2 >= dy {
			err += dy // Adjust the error tracker to account for the horizontal step.
			x0 += sx  // Take one step along the X-axis.
		}

		// Adjust Y if our drift tolerance allows it.
		// Note: Both if-statements can trigger in the same iteration, resulting in a diagonal step.
		if e2 <= dx {
			err += dx // Adjust the error tracker to account for the vertical step.
			y0 += sy  // Take one step along the Y-axis.
		}
	}

	// Return the complete charted path to the caller.
	return points
}
