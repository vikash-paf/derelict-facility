package systems

import (
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// RenderEntities loops through all entities possessing BOTH a Sprite and Position component
// and draws them to the active display buffer if they are within exactly visible map tiles.
func RenderEntities(w *ecs.World, disp display.Display, gameMap *world.Map) {
	// A more robust ECS would have a function to get entities with multiple components,
	// but for simplicity we iterate over one and check for the other.
	entities := w.GetEntitiesWith(components.NameSprite)
	for _, e := range entities {
		renderRaw := w.GetComponent(e, components.NameSprite)
		posRaw := w.GetComponent(e, components.NamePosition)

		if renderRaw == nil || posRaw == nil {
			continue // Need both to draw!
		}

		renderCfg := renderRaw.(*components.Sprite)
		pos := posRaw.(*components.Position)

		// Respect the fog of war! Don't draw entities we can't see.
		// Exception: the player themself should always be drawn, even if standing in a glitched dark tile
		isPlayer := w.GetComponent(e, components.NamePlayerControl) != nil
		if !isPlayer {
			tile := gameMap.GetTile(pos.X, pos.Y)
			if tile == nil || !tile.Visible {
				continue
			}
		}

		hexColor := uint32(0xFFFFFFFF)
		if renderCfg.ColorCode != "" {
			hexColor = display.MapANSIColor(renderCfg.ColorCode)
		}

		disp.DrawSprite(pos.X, pos.Y, renderCfg.SheetX, renderCfg.SheetY, hexColor)
	}
}
