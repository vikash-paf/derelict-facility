package systems

import (
	"math/rand"

	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// ProcessAutopilot handles AI pathing and automated door opening for PlayerControl entities.
func ProcessAutopilot(w *ecs.World, gameMap *world.Map, pf *world.Pathfinder, audioFunc func(string)) {
	targetMask := components.MaskPlayerControl | components.MaskPosition

	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & targetMask) != targetMask {
			continue
		}

		ctrl := &w.PlayerControls[i]
		pos := &w.Positions[i]

		if !ctrl.Autopilot {
			continue
		}

		// If no active path, attempt to find a path to a random room
		if len(ctrl.CurrentPath) == 0 {
			ctrl.CurrentPath = findRandomRoomPath(w, gameMap, pf, *pos, ctrl.SecurityClearance)
			if len(ctrl.CurrentPath) == 0 {
				ctrl.Autopilot = false
			}
			continue
		}

		// Take the next step along the path
		nextStep := ctrl.CurrentPath[0]
		tryAutoOpenDoor(w, nextStep, ctrl.SecurityClearance, audioFunc)

		if !gameMap.IsWalkable(nextStep.X, nextStep.Y) || IsSolidAt(w, nextStep.X, nextStep.Y) {
			ctrl.CurrentPath = nil // Path blocked, recalculate on next tick
			continue
		}

		pos.X, pos.Y = nextStep.X, nextStep.Y
		ctrl.CurrentPath = ctrl.CurrentPath[1:]

		if audioFunc != nil {
			audioFunc("footstep")
		}
	}
}

// findRandomRoomPath shuffles and searches for any reachable room, choosing a walkable target tile within it.
func findRandomRoomPath(
	w *ecs.World,
	gameMap *world.Map,
	pf *world.Pathfinder,
	startPos components.Position,
	clearance uint32,
) []entity.Point {
	if len(gameMap.Rooms) == 0 {
		return nil
	}

	start := entity.Point{X: startPos.X, Y: startPos.Y}

	// Shuffle rooms to ensure randomness without giving up early
	shuffledRooms := make([]world.Rect, len(gameMap.Rooms))
	copy(shuffledRooms, gameMap.Rooms)
	rand.Shuffle(len(shuffledRooms), func(i, j int) {
		shuffledRooms[i], shuffledRooms[j] = shuffledRooms[j], shuffledRooms[i]
	})

	for _, room := range shuffledRooms {
		var walkablePoints []entity.Point
		for rx := room.X1; rx <= room.X2; rx++ {
			for ry := room.Y1; ry <= room.Y2; ry++ {
				if gameMap.IsWalkable(rx, ry) && isTileTraversable(w, gameMap, rx, ry, clearance) {
					walkablePoints = append(walkablePoints, entity.Point{X: rx, Y: ry})
				}
			}
		}

		if len(walkablePoints) == 0 {
			continue
		}

		// Choose a random walkable target inside the selected room
		target := walkablePoints[rand.Intn(len(walkablePoints))]
		path := pf.FindPath(gameMap, start, target, func(x, y int) bool {
			return isTileTraversable(w, gameMap, x, y, clearance)
		})

		if len(path) > 1 {
			return path[1:]
		}
	}

	return nil
}

// isTileTraversable determines if a coordinate is walkable or is an unlocked door.
func isTileTraversable(w *ecs.World, gameMap *world.Map, x, y int, clearance uint32) bool {
	if !gameMap.IsWalkable(x, y) {
		return false
	}

	doorMask := components.MaskPosition | components.MaskDoor
	solidMask := components.MaskPosition | components.MaskSolid

	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		mask := w.Masks[i]
		if (mask & solidMask) != solidMask {
			continue
		}

		pos := w.Positions[i]
		if pos.X != x || pos.Y != y {
			continue
		}

		// If solid entity is a door, check if player has clearance to unlock it
		if (mask & doorMask) == doorMask {
			door := w.Doors[i]
			return door.RequiredClearance == 0 || (clearance&door.RequiredClearance) != 0
		}

		return false // Wall or obstacle
	}

	return true
}

// tryAutoOpenDoor automatically unlocks and opens closed doors when encountered on autopilot.
func tryAutoOpenDoor(w *ecs.World, step entity.Point, clearance uint32, audioFunc func(string)) {
	doorMask := components.MaskPosition | components.MaskDoor

	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (w.Masks[i] & doorMask) != doorMask {
			continue
		}

		pos := w.Positions[i]
		if pos.X != step.X || pos.Y != step.Y {
			continue
		}

		door := &w.Doors[i]
		if door.IsOpen {
			return
		}

		// Check security clearance requirement
		if door.RequiredClearance != 0 && (clearance&door.RequiredClearance) == 0 {
			return
		}

		door.IsOpen = true
		w.RemoveSolid(i)
		w.Interactables[i].Prompt = "Press [E] to Close Door"

		if (w.Masks[i] & components.MaskGlyph) != 0 {
			w.Glyphs[i].Char = "/"
			w.Glyphs[i].Color = core.Gray
		}

		if audioFunc != nil {
			audioFunc("door_open")
		}
		return
	}
}
