package systems

import (
	"fmt"

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
func ProcessPlayerInput(w *ecs.World, events []core.InputEvent, gameMap *world.Map, logFunc func(string), audioFunc func(string), transitionFunc func(string, bool)) {
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
				handleInteraction(w, positions.X, positions.Y, gameMap, logFunc, audioFunc, transitionFunc)
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
					if audioFunc != nil {
						audioFunc("footstep")
					}
				}
			}
		}
	}
}

func handleInteraction(w *ecs.World, playerX, playerY int, gameMap *world.Map, logFunc func(string), audioFunc func(string), transitionFunc func(string, bool)) {
	// Find the player's clearance first
	playerClearance := uint32(0)
	playerEntID := ecs.Entity(0)
	foundPlayer := false
	playerMask := components.MaskPlayerControl
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & playerMask) == playerMask {
			playerClearance = w.PlayerControls[i].SecurityClearance
			playerEntID = i
			foundPlayer = true
			break
		}
	}

	targetMask := components.MaskPosition | components.MaskInteractable
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & targetMask) == targetMask {
			pos := w.Positions[i]
			// Check adjacency
			dx := pos.X - playerX
			dy := pos.Y - playerY
			distSq := dx*dx + dy*dy

			if distSq <= 2 {
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
					if audioFunc != nil {
						audioFunc("generator_toggle")
					}
					logFunc("Power Generator toggled.")
					return
				}

				// 2. Door
				if (w.Masks[i] & components.MaskDoor) != 0 {
					door := &w.Doors[i]

					// Check security clearance
					if door.RequiredClearance != 0 && (playerClearance&door.RequiredClearance) == 0 {
						if audioFunc != nil {
							audioFunc("access_denied")
						}
						logFunc("ACCESS DENIED: Required Clearance Missing.")
						return
					}

					door.IsOpen = !door.IsOpen

					if door.IsOpen {
						w.RemoveSolid(i)
						w.Interactables[i].Prompt = "Press [E] to Close Door"
						if (w.Masks[i] & components.MaskGlyph) != 0 {
							w.Glyphs[i].Char = "/"
							w.Glyphs[i].Color = core.Gray
						}
						if audioFunc != nil {
							audioFunc("door_open")
						}
						logFunc("Door opened.")
					} else {
						w.AddSolid(i)
						w.Interactables[i].Prompt = "Press [E] to Open Door"
						if (w.Masks[i] & components.MaskGlyph) != 0 {
							w.Glyphs[i].Char = "+"
							w.Glyphs[i].Color = core.White
						}
						if audioFunc != nil {
							audioFunc("door_close")
						}
						logFunc("Door closed.")
					}
					return
				}

				// 3. Terminal
				if (w.Masks[i] & components.MaskTerminal) != 0 {
					terminal := &w.Terminals[i]

					if audioFunc != nil {
						audioFunc("terminal_access")
					}

					// Grant clearance if terminal has it
					if terminal.GrantClearance != 0 && foundPlayer {
						if (w.PlayerControls[playerEntID].SecurityClearance & terminal.GrantClearance) == 0 {
							w.PlayerControls[playerEntID].SecurityClearance |= terminal.GrantClearance
							logFunc("SECURITY CLEARANCE UPDATED.")
						}
					}

					// If it has narrative data, display it
					if (w.Masks[i] & components.MaskNarrative) != 0 {
						logFunc(w.Narratives[i].Text)
					}

					if !terminal.HasSaved {
						terminal.HasSaved = true
						w.Interactables[i].Prompt = "Press [E] to Access Terminal"
						if (w.Masks[i] & components.MaskGlyph) != 0 {
							w.Glyphs[i].Color = core.Green
						}
						saveState(w, gameMap)
						logFunc("Checkpoint saved.")
					}
					return
				}

				// 4. Stairway / Elevator Level Transition
				if (w.Masks[i] & components.MaskStairway) != 0 {
					stair := &w.Stairways[i]
					if stair.RequiredClearance != 0 && (playerClearance&stair.RequiredClearance) == 0 {
						if audioFunc != nil {
							audioFunc("access_denied")
						}
						logFunc("ELEVATOR LOCKED: Required Clearance Missing.")
						return
					}
					if stair.TargetLevelID == "" {
						logFunc("No connecting level found.")
						return
					}
					if audioFunc != nil {
						audioFunc("terminal_access")
					}
					dirLabel := "Descending"
					if stair.IsUp {
						dirLabel = "Ascending"
					}
					logFunc(fmt.Sprintf("ELEVATOR: %s to %s...", dirLabel, stair.TargetLevelID))
					if transitionFunc != nil {
						transitionFunc(stair.TargetLevelID, stair.IsUp)
					}
					return
				}
			}
		}
	}
}

// IsPowerActive returns true if at least one global generator is active.
func IsPowerActive(w *ecs.World) bool {
	targetMask := components.MaskPowerGenerator
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & targetMask) == targetMask {
			gen := w.PowerGenerators[i]
			if gen.IsActive && gen.IsGlobal {
				return true
			}
		}
	}
	return false
}

// IsPowerActiveAt returns true if the global power is active OR if an active local generator powers the room containing (x,y).
func IsPowerActiveAt(w *ecs.World, gameMap *world.Map, x, y int) bool {
	if IsPowerActive(w) {
		return true
	}

	targetMask := components.MaskPosition | components.MaskPowerGenerator
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & targetMask) == targetMask {
			gen := w.PowerGenerators[i]
			if !gen.IsActive || gen.IsGlobal {
				continue
			}

			// Local generator: check if the generator is in the same room as (x,y)
			genPos := w.Positions[i]
			for _, room := range gameMap.Rooms {
				if room.Contains(genPos.X, genPos.Y) && room.Contains(x, y) {
					return true
				}
			}
		}
	}

	return false
}
