package systems

import (
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// ProcessLifeSupport updates the player's survival components based on powered and sunlit locations.
func ProcessLifeSupport(w *ecs.World, gameMap *world.Map, clock *world.FacilityClock, logFunc func(string)) {
	for i := range ecs.Entity(ecs.MaxEntities) {
		if w.IsPlayer(i) && w.HasPosition(i) {
			pos := w.Positions[i]
			playerCtrl := &w.PlayerControls[i]
			survival := &playerCtrl.Survival

			// Check if current tile is powered or sunlit during daytime
			isPowered := IsPowerActiveAt(w, gameMap, pos.X, pos.Y)
			tile := gameMap.GetTile(pos.X, pos.Y)
			isSunlit := tile != nil && tile.SunlightIntensity > 0.0 && clock.IsDaytime()

			if isPowered || isSunlit {
				// Breathing safely: restore O2 and clean toxins
				survival.Oxygen = min(survival.MaxOxygen, survival.Oxygen+0.4)
				if survival.Toxicity > 0 {
					survival.Toxicity = max(0.0, survival.Toxicity-0.1)
				}
				if playerCtrl.Status == components.PlayerStatusSuffocating {
					playerCtrl.Status = components.PlayerStatusHealthy
				}
			} else {
				// Dangerous dark/unpowered zone: deplete O2
				survival.Oxygen = max(0.0, survival.Oxygen-0.15)
				if survival.Oxygen <= 0.0 {
					// Suffocation sets in: toxicity grows and health drops
					survival.Toxicity = min(100.0, survival.Toxicity+0.5)
					survival.Health = max(0.0, survival.Health-0.4)
					playerCtrl.Status = components.PlayerStatusSuffocating
				} else {
					if playerCtrl.Status == components.PlayerStatusSuffocating {
						playerCtrl.Status = components.PlayerStatusHealthy
					}
				}
			}

			// Apply health decay if heavily toxic
			if survival.Toxicity >= 50.0 {
				survival.Health = max(0.0, survival.Health-0.1)
				if playerCtrl.Status != components.PlayerStatusSuffocating {
					playerCtrl.Status = components.PlayerStatusSick
				}
			} else if playerCtrl.Status == components.PlayerStatusSick {
				playerCtrl.Status = components.PlayerStatusHealthy
			}

			// Health status styling
			if survival.Health < 30.0 && playerCtrl.Status == components.PlayerStatusHealthy {
				playerCtrl.Status = components.PlayerStatusHurt
			}
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
