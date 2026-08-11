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
				survival.Oxygen = min(survival.MaxOxygen, survival.Oxygen+0.6)
				if survival.Toxicity > 0 {
					// Faster toxin clearing when breathing safely
					survival.Toxicity = max(0.0, survival.Toxicity-0.2)
				}
				if playerCtrl.Status == components.PlayerStatusSuffocating {
					playerCtrl.Status = components.PlayerStatusHealthy
				}

				// Passive health recovery: regenerate health slowly (+1 HP per 3 ticks) in safe zones if clean of toxins
				if survival.Toxicity == 0.0 && survival.Health < survival.MaxHealth {
					survival.Health = min(survival.MaxHealth, survival.Health+0.35)
				}
			} else {
				// Dangerous dark/unpowered zone: deplete O2 more slowly (0.08 down from 0.15)
				survival.Oxygen = max(0.0, survival.Oxygen-0.08)
				if survival.Oxygen <= 0.0 {
					// Suffocation sets in: toxicity grows and health drops slightly slower (0.20 down from 0.4)
					survival.Toxicity = min(100.0, survival.Toxicity+0.3)
					survival.Health = max(0.0, survival.Health-0.2)
					playerCtrl.Status = components.PlayerStatusSuffocating
				} else {
					if playerCtrl.Status == components.PlayerStatusSuffocating {
						playerCtrl.Status = components.PlayerStatusHealthy
					}
				}
			}

			// Apply health decay if heavily toxic (reduced from 0.1 to 0.05 per tick)
			if survival.Toxicity >= 50.0 {
				survival.Health = max(0.0, survival.Health-0.05)
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
