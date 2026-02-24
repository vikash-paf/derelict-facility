package systems

import (
	"math/rand"

	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// ProcessAutopilot handles the AI pathing logic for any Entity with PlayerControl.
func ProcessAutopilot(w *ecs.World, gameMap *world.Map, pf *world.Pathfinder) {
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
				// Pick a random room
				targetRoom := gameMap.Rooms[rand.Intn(len(gameMap.Rooms))]
				targetX, targetY := targetRoom.Center()

				start := entity.Point{X: pos.X, Y: pos.Y}
				target := entity.Point{X: targetX, Y: targetY}

				// Calculate the path
				path := pf.FindPath(gameMap, start, target)

				if len(path) > 1 {
					ctrl.CurrentPath = path[1:]
				} else {
					ctrl.CurrentPath = nil // Already there
				}
				continue
			}

			// 2. Take the next step in the path
			nextStep := ctrl.CurrentPath[0]

			if gameMap.IsWalkable(nextStep.X, nextStep.Y) {
				pos.X = nextStep.X
				pos.Y = nextStep.Y
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
