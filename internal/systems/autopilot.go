package systems

import (
	"math/rand"

	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// ProcessAutopilot handles the AI pathing logic for any Entity with PlayerControl.
func ProcessAutopilot(w *ecs.World, gameMap *world.Map, pf *world.Pathfinder, audioFunc func(string)) {
	targetMask := components.MaskPlayerControl | components.MaskPosition

	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & targetMask) == targetMask {
			ctrl := &w.PlayerControls[i]
			pos := &w.Positions[i]

			if !ctrl.Autopilot {
				continue // AI is toggled off
			}

			// 1. If we don't have a path, find a new destination!
			if len(ctrl.CurrentPath) == 0 {
				if len(gameMap.Rooms) == 0 {
					// Fallback if no rooms are defined
					ctrl.Autopilot = false
					continue
				}
				// Try up to 5 random rooms to find one with a valid path
				for attempt := 0; attempt < 5; attempt++ {
					targetRoom := gameMap.Rooms[rand.Intn(len(gameMap.Rooms))]
					targetX, targetY := targetRoom.Center()

					start := entity.Point{X: pos.X, Y: pos.Y}
					target := entity.Point{X: targetX, Y: targetY}

					// Calculate the path (doors count as walkable if unlocked)
					path := pf.FindPath(gameMap, start, target, func(x, y int) bool {
						if !gameMap.IsWalkable(x, y) {
							return false
						}
						// Solid check: if there is a solid entity, check if it's an unlocked door
						targetMask := components.MaskPosition | components.MaskSolid
						for eID := ecs.Entity(0); eID < ecs.MaxEntities; eID++ {
							if (w.Masks[eID] & targetMask) == targetMask {
								solidPos := w.Positions[eID]
								if solidPos.X == x && solidPos.Y == y {
									// Check if it's a door
									if (w.Masks[eID] & components.MaskDoor) != 0 {
										door := w.Doors[eID]
										if door.RequiredClearance == 0 || (ctrl.SecurityClearance&door.RequiredClearance) != 0 {
											return true // Autopilot can open unlocked doors!
										}
									}
									return false // Solid obstacle or locked door
								}
							}
						}
						return true
					})

					if len(path) > 1 {
						ctrl.CurrentPath = path[1:]
						break
					}
				}
				if len(ctrl.CurrentPath) == 0 {
					ctrl.Autopilot = false
				}
				continue
			}

			// 2. Take the next step in the path
			nextStep := ctrl.CurrentPath[0]

			// Check if next step is a closed door and auto-open it
			doorMask := components.MaskPosition | components.MaskDoor
			for eID := ecs.Entity(0); eID < ecs.MaxEntities; eID++ {
				if (w.Masks[eID] & doorMask) == doorMask {
					doorPos := w.Positions[eID]
					if doorPos.X == nextStep.X && doorPos.Y == nextStep.Y {
						door := &w.Doors[eID]
						if !door.IsOpen {
							if door.RequiredClearance == 0 || (ctrl.SecurityClearance&door.RequiredClearance) != 0 {
								door.IsOpen = true
								w.RemoveSolid(eID)
								w.Interactables[eID].Prompt = "Press [E] to Close Door"
								if (w.Masks[eID] & components.MaskGlyph) != 0 {
									w.Glyphs[eID].Char = "/"
									w.Glyphs[eID].Color = core.Gray
								}
								if audioFunc != nil {
									audioFunc("door_open")
								}
							}
						}
						break
					}
				}
			}

			if gameMap.IsWalkable(nextStep.X, nextStep.Y) && !IsSolidAt(w, nextStep.X, nextStep.Y) {
				pos.X = nextStep.X
				pos.Y = nextStep.Y
				if audioFunc != nil {
					audioFunc("footstep")
				}
			} else {
				// Path is blocked! Clear it so we recalculate next tick.
				ctrl.CurrentPath = nil
				continue
			}

			// 3. Pop the step we just took off the slice
			ctrl.CurrentPath = ctrl.CurrentPath[1:]
		}
	}
}
