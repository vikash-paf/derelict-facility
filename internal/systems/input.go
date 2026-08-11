package systems

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/mission"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// IsSolidAt checks if any solid entity occupies the given coordinates.
func IsSolidAt(w *ecs.World, x, y int) bool {
	for i := range ecs.Entity(ecs.MaxEntities) {
		if w.HasPosition(i) && w.IsSolid(i) {
			pos := w.Positions[i]
			if pos.X == x && pos.Y == y {
				return true
			}
		}
	}
	return false
}

// ProcessPlayerInput handles intentional movement from W/A/S/D.
func ProcessPlayerInput(w *ecs.World, events []core.InputEvent, gameMap *world.Map, activeMission *mission.MissionManifest, activeLevelID string, logFunc func(string), audioFunc func(string), transitionFunc func(string, bool), saveFunc func()) {
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

	for i := range ecs.Entity(ecs.MaxEntities) {
		if w.IsPlayer(i) && w.HasPosition(i) {
			controls := &w.PlayerControls[i]
			positions := &w.Positions[i]

			if toggleAutopilot {
				controls.Autopilot = !controls.Autopilot
				controls.CurrentPath = nil // clear path when toggling
			}

			if interactPressed {
				// Find adjacent interactable entities
				handleInteraction(w, positions.X, positions.Y, gameMap, activeMission, activeLevelID, logFunc, audioFunc, transitionFunc, saveFunc)
			}

			// Disable autopilot if manual movement keys are pressed
			if dx != 0 || dy != 0 {
				if controls.Autopilot {
					controls.Autopilot = false
					controls.CurrentPath = nil
				}
			}

			if dx == 0 && dy == 0 {
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

func interactWithGenerator(w *ecs.World, ent ecs.Entity, logFunc func(string), audioFunc func(string)) {
	gen := &w.PowerGenerators[ent]
	gen.IsActive = !gen.IsActive

	// Update visual feedback and interaction prompt
	if (w.Masks[ent] & components.MaskGlyph) != 0 {
		glyph := &w.Glyphs[ent]
		if gen.IsActive {
			glyph.Color = core.Green
			glyph.Char = "⚡"
		} else {
			glyph.Color = core.Red
			glyph.Char = "X"
		}
	}

	if gen.IsActive {
		w.Interactables[ent].Prompt = "Press [E] to Turn Off Generator"
	} else {
		w.Interactables[ent].Prompt = "Press [E] to Turn On Generator"
	}

	if audioFunc != nil {
		audioFunc("generator_toggle")
	}

	if gen.IsActive {
		logFunc("Power Generator turned on.")
	} else {
		logFunc("Power Generator turned off.")
	}
}

func interactWithDoor(w *ecs.World, ent ecs.Entity, playerClearance uint32, logFunc func(string), audioFunc func(string)) {
	door := &w.Doors[ent]

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
		w.RemoveSolid(ent)
		w.Interactables[ent].Prompt = "Press [E] to Close Door"
		if (w.Masks[ent] & components.MaskGlyph) != 0 {
			w.Glyphs[ent].Char = "/"
			w.Glyphs[ent].Color = core.Gray
		}
		if audioFunc != nil {
			audioFunc("door_open")
		}
		logFunc("Door opened.")
	} else {
		w.AddSolid(ent)
		w.Interactables[ent].Prompt = "Press [E] to Open Door"
		if (w.Masks[ent] & components.MaskGlyph) != 0 {
			w.Glyphs[ent].Char = "+"
			w.Glyphs[ent].Color = core.White
		}
		if audioFunc != nil {
			audioFunc("door_close")
		}
		logFunc("Door closed.")
	}
}

func interactWithTerminal(w *ecs.World, ent ecs.Entity, playerEntID ecs.Entity, foundPlayer bool, gameMap *world.Map, activeMission *mission.MissionManifest, activeLevelID string, logFunc func(string), audioFunc func(string), saveFunc func()) {
	terminal := &w.Terminals[ent]

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
	if (w.Masks[ent] & components.MaskNarrative) != 0 {
		logFunc(w.Narratives[ent].Text)
	}

	if !terminal.HasSaved {
		terminal.HasSaved = true
		w.Interactables[ent].Prompt = "Press [E] to Access Terminal"
		if (w.Masks[ent] & components.MaskGlyph) != 0 {
			w.Glyphs[ent].Color = core.Green
		}
		if saveFunc != nil {
			saveFunc()
		}
		logFunc("Checkpoint saved.")
	}
}

func interactWithStairway(w *ecs.World, ent ecs.Entity, playerClearance uint32, logFunc func(string), audioFunc func(string), transitionFunc func(string, bool)) {
	stair := &w.Stairways[ent]
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
}

func handleInteraction(
	w *ecs.World,
	playerX, playerY int,
	gameMap *world.Map,
	activeMission *mission.MissionManifest,
	activeLevelID string,
	logFunc func(string),
	audioFunc func(string),
	transitionFunc func(string, bool),
	saveFunc func(),
) {
	// Find the player's clearance first
	playerClearance := uint32(0)
	playerEntID := ecs.Entity(0)
	foundPlayer := false
	for i := range ecs.Entity(ecs.MaxEntities) {
		if w.IsPlayer(i) {
			playerClearance = w.PlayerControls[i].SecurityClearance
			playerEntID = i
			foundPlayer = true
			break
		}
	}

	for i := range ecs.Entity(ecs.MaxEntities) {
		if w.HasPosition(i) && w.IsInteractable(i) {
			pos := w.Positions[i]
			// Check adjacency
			dx := pos.X - playerX
			dy := pos.Y - playerY
			distSq := dx*dx + dy*dy

			if distSq <= 2 {
				switch {
				case w.IsPowerGenerator(i):
					interactWithGenerator(w, i, logFunc, audioFunc)
					return
				case w.IsDoor(i):
					interactWithDoor(w, i, playerClearance, logFunc, audioFunc)
					return
				case w.IsTerminal(i):
					interactWithTerminal(w, i, playerEntID, foundPlayer, gameMap, activeMission, activeLevelID, logFunc, audioFunc, saveFunc)
					return
				case w.IsStairway(i):
					interactWithStairway(w, i, playerClearance, logFunc, audioFunc, transitionFunc)
					return
				}
			}
		}
	}

	// Check if player is standing on or adjacent to a harvestable map-tile plant
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			tx := playerX + dx
			ty := playerY + dy
			tile := gameMap.GetTile(tx, ty)
			if tile != nil && tile.PlantStage == world.PlantStageMature && foundPlayer {
				playerCtrl := &w.PlayerControls[playerEntID]
				if tile.YieldItemType == "O2_CAPSULE" {
					playerCtrl.Survival.Oxygen = min(playerCtrl.Survival.MaxOxygen, playerCtrl.Survival.Oxygen+40.0)
					logFunc("Harvested plant: Restored 40% Oxygen.")
				} else if tile.YieldItemType == "MEDPACK" {
					playerCtrl.Survival.Toxicity = max(0.0, playerCtrl.Survival.Toxicity-30.0)
					playerCtrl.Survival.Health = min(playerCtrl.Survival.MaxHealth, playerCtrl.Survival.Health+25.0)
					logFunc("Harvested plant: Treated toxicity (-30%) and healed 25 HP.")
				}

				if audioFunc != nil {
					audioFunc("generator_toggle")
				}

				// Reset tile plant state
				tile.PlantStage = world.PlantStageSeed
				tile.GrowthProgress = 0.0
				return
			}
		}
	}
}

// IsPowerActive returns true if at least one global generator is active.
func IsPowerActive(w *ecs.World) bool {
	for i := range ecs.Entity(ecs.MaxEntities) {
		if w.IsPowerGenerator(i) {
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

	for i := range ecs.Entity(ecs.MaxEntities) {
		if w.HasPosition(i) && w.IsPowerGenerator(i) {
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
