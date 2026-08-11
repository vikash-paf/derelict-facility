package systems

import (
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// ProcessHydroponics simulates plant growth depending on tile light & power configurations.
func ProcessHydroponics(w *ecs.World, gameMap *world.Map, clock *world.FacilityClock) {
	for i := range ecs.Entity(ecs.MaxEntities) {
		if w.IsHydroponics(i) && w.HasPosition(i) {
			pos := w.Positions[i]
			plant := &w.Hydroponics[i]

			if plant.Stage == components.PlantStageMature {
				continue
			}

			// Calculate sunlight factor at the plant's tile
			tile := gameMap.GetTile(pos.X, pos.Y)
			sunlightIntensity := 0.0
			if tile != nil && tile.IsSunlit && clock.IsDaytime() {
				sunlightIntensity = tile.SunlightIntensity
			}

			// Calculate local room power factor
			isPowered := IsPowerActiveAt(w, gameMap, pos.X, pos.Y)
			powerFactor := 0.2 // basic battery backup
			if isPowered {
				powerFactor = 1.0
			}

			// Growth rate is combined sunlight and power
			growthAmount := plant.GrowthRate * ((sunlightIntensity * 0.7) + (powerFactor * 0.3))
			plant.GrowthProgress += growthAmount

			if plant.GrowthProgress >= 100.0 {
				plant.GrowthProgress = 0.0
				if plant.Stage == components.PlantStageSeed {
					plant.Stage = components.PlantStageSprout
					if (w.Masks[i] & components.MaskGlyph) != 0 {
						w.Glyphs[i].Char = "v"
						w.Glyphs[i].Color = core.Color{R: 120, G: 200, B: 120, A: 255}
					}
				} else if plant.Stage == components.PlantStageSprout {
					plant.Stage = components.PlantStageMature
					w.AddInteractable(i, components.Interactable{Prompt: "Press [E] to Harvest Plant"})
					if (w.Masks[i] & components.MaskGlyph) != 0 {
						w.Glyphs[i].Char = "♣"
						w.Glyphs[i].Color = core.Color{R: 50, G: 240, B: 50, A: 255}
					}
				}
			}
		}
	}
}
