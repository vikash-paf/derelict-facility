package systems

import (
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// ProcessPlayerInput handles intentional movement from W/A/S/D.
func ProcessPlayerInput(w *ecs.World, events []core.InputEvent, gameMap *world.Map) {
	// First analyze events to see if we pressed WASD or P
	dx, dy := 0, 0
	toggleAutopilot := false

	for _, event := range events {
		switch event.Key {
		case core.KeyW:
			dy = -1
		case core.KeyS:
			dy = 1
		case core.KeyA:
			dx = -1
		case core.KeyD:
			dx = 1
		case core.KeyP:
			toggleAutopilot = true
		}
	}

	// Now apply to any Entity with a "PlayerControl" and a "Position" component.
	entities := w.GetEntitiesWith(components.NamePlayerControl)
	for _, e := range entities {
		ctrlRaw := w.GetComponent(e, components.NamePlayerControl)
		posRaw := w.GetComponent(e, components.NamePosition)

		if ctrlRaw == nil || posRaw == nil {
			continue // Entity is malformed
		}

		ctrl := ctrlRaw.(*components.PlayerControl)
		pos := posRaw.(*components.Position)

		if toggleAutopilot {
			ctrl.Autopilot = !ctrl.Autopilot
			ctrl.CurrentPath = nil // clear path when toggling
		}

		// Don't manually move if Autopilot is running
		if ctrl.Autopilot || (dx == 0 && dy == 0) {
			continue
		}

		newX := pos.X + dx
		newY := pos.Y + dy

		// ensure valid move
		if newX >= 0 && newX < gameMap.Width && newY >= 0 && newY < gameMap.Height {
			tile := gameMap.GetTile(newX, newY)
			if tile != nil && tile.Walkable {
				pos.X = newX
				pos.Y = newY
			}
		}
	}
}
