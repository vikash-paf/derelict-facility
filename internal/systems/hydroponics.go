package systems

import (
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// ProcessHydroponics simulates plant growth directly inside the grid's tiles.
func ProcessHydroponics(w *ecs.World, gameMap *world.Map, clock *world.FacilityClock) {
	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			tile := gameMap.GetTile(x, y)
			if tile == nil || tile.PlantStage == world.PlantStageNone || tile.PlantStage == world.PlantStageMature {
				continue
			}

			// Calculate sunlight intensity
			sunlightIntensity := 0.0
			if tile.IsSunlit && clock.IsDaytime() {
				sunlightIntensity = tile.SunlightIntensity
			}

			// Calculate local room power factor
			isPowered := IsPowerActiveAt(w, gameMap, x, y)
			powerFactor := 0.2 // basic battery backup
			if isPowered {
				powerFactor = 1.0
			}

			// Growth rate is combined sunlight and power
			growthAmount := tile.GrowthRate * ((sunlightIntensity * 0.7) + (powerFactor * 0.3))
			tile.GrowthProgress += growthAmount

			if tile.GrowthProgress >= 100.0 {
				tile.GrowthProgress = 0.0
				if tile.PlantStage == world.PlantStageSeed {
					tile.PlantStage = world.PlantStageSprout
				} else if tile.PlantStage == world.PlantStageSprout {
					tile.PlantStage = world.PlantStageMature
				}
			}
		}
	}
}
