package systems

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// IsSolidAt checks if any solid entity occupies the given coordinates.
func IsSolidAt(w *ecs.World, x, y int) bool {
	targetMask := components.MaskPosition | components.MaskSolid
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & targetMask) == targetMask {
			pos := w.Positions[i]
			if pos.X == x && pos.Y == y {
				return true
			}
		}
	}
	return false
}

// ProcessPlayerInput handles intentional movement from W/A/S/D.
func ProcessPlayerInput(w *ecs.World, events []core.InputEvent, gameMap *world.Map) {
	dx, dy := 0, 0
	toggleAutopilot := false
	interactPressed := false

	for _, event := range events {
		switch event.Key {
		case rl.KeyW:
			dy = -1
		case rl.KeyS:
			dy = 1
		case rl.KeyA:
			dx = -1
		case rl.KeyD:
			dx = 1
		case rl.KeyP:
			toggleAutopilot = true
		case rl.KeyE:
			interactPressed = true
		}
	}

	targetMask := components.MaskPlayerControl | components.MaskPosition

	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & targetMask) == targetMask {
			controls := &w.PlayerControls[i]
			positions := &w.Positions[i]

			if toggleAutopilot {
				controls.Autopilot = !controls.Autopilot
				controls.CurrentPath = nil // clear path when toggling
			}

			if interactPressed {
				// Find adjacent interactable entities
				handleInteraction(w, positions.X, positions.Y, gameMap)
			}

			// Don't manually move if Autopilot is running
			if controls.Autopilot || (dx == 0 && dy == 0) {
				continue
			}

			newX := positions.X + dx
			newY := positions.Y + dy

			// ensure valid move
			if newX >= 0 && newX < gameMap.Width && newY >= 0 && newY < gameMap.Height {
				tile := gameMap.GetTile(newX, newY)
				if tile != nil && tile.Walkable && !IsSolidAt(w, newX, newY) {
					positions.X = newX
					positions.Y = newY
				}
			}
		}
	}
}

func handleInteraction(w *ecs.World, playerX, playerY int, gameMap *world.Map) {
	targetMask := components.MaskPosition | components.MaskInteractable
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & targetMask) == targetMask {
			pos := w.Positions[i]
			// Check adjacency (including diagonals, or just orthogonal?)
			// Orthogonal:
			dx := pos.X - playerX
			dy := pos.Y - playerY
			distSq := dx*dx + dy*dy

			if distSq <= 2 { // 1 tile away orthogonally (distSq=1) or diagonally (distSq=2) or same tile (0)
				// What kind of interactable is it?

				// 1. Power Generator
				if (w.Masks[i] & components.MaskPowerGenerator) != 0 {
					gen := &w.PowerGenerators[i]
					gen.IsActive = !gen.IsActive

					// Update visual feedback
					if (w.Masks[i] & components.MaskGlyph) != 0 {
						glyph := &w.Glyphs[i]
						if gen.IsActive {
							glyph.Color = core.Green
							glyph.Char = "⚡"
						} else {
							glyph.Color = core.Red
							glyph.Char = "X"
						}
					}
					return // Stop after interacting
				}

				// 2. Door
				if (w.Masks[i] & components.MaskDoor) != 0 {
					door := &w.Doors[i]
					door.IsOpen = !door.IsOpen

					if door.IsOpen {
						// Open the door
						w.RemoveSolid(i)
						w.Interactables[i].Prompt = "Press [E] to Close Door"
						if (w.Masks[i] & components.MaskGlyph) != 0 {
							w.Glyphs[i].Char = "/"
							w.Glyphs[i].Color = core.Gray
						}
					} else {
						// Close the door
						w.AddSolid(i)
						w.Interactables[i].Prompt = "Press [E] to Open Door"
						if (w.Masks[i] & components.MaskGlyph) != 0 {
							w.Glyphs[i].Char = "+"
							w.Glyphs[i].Color = core.White
						}
					}
					return // Stop after interacting
				}

				// 3. Terminal
				if (w.Masks[i] & components.MaskTerminal) != 0 {
					terminal := &w.Terminals[i]
					if !terminal.HasSaved {
						terminal.HasSaved = true
						w.Interactables[i].Prompt = "[ CHECKPOINT SAVED ]"
						if (w.Masks[i] & components.MaskGlyph) != 0 {
							w.Glyphs[i].Color = core.Green
						}
						saveState(w, gameMap)
					}
					return // Stop after interacting
				}
			}
		}
	}
}

// IsPowerActive returns true if at least one generator is currently active
func IsPowerActive(w *ecs.World) bool {
	targetMask := components.MaskPowerGenerator
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & targetMask) == targetMask {
			if w.PowerGenerators[i].IsActive {
				return true
			}
		}
	}
	return false
}
