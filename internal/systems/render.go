package systems

import (
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// RenderEntities loops through all entities possessing BOTH a Sprite or Glyph and Position component
// and draws them to the active display buffer if they are within exactly visible map tiles.
func RenderEntities(w *ecs.World, disp display.Display, gameMap *world.Map) {
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		// Must have a position to be rendered
		if (w.Masks[i] & components.MaskPosition) == 0 {
			continue
		}

		hasSprite := (w.Masks[i] & components.MaskSprite) != 0
		hasGlyph := (w.Masks[i] & components.MaskGlyph) != 0

		// Must have at least one visual representation
		if !hasSprite && !hasGlyph {
			continue
		}

		pos := w.Positions[i]

		// Is it the player? Check the mask for the PlayerControl bit
		isPlayer := (w.Masks[i] & components.MaskPlayerControl) != 0

		if !isPlayer {
			tile := gameMap.GetTile(pos.X, pos.Y)
			if tile == nil || !tile.Visible {
				continue
			}
		}

		if hasSprite {
			spr := w.Sprites[i]
			hexColor := uint32(0xFFFFFFFF)
			if spr.ColorCode != "" {
				hexColor = display.MapANSIColor(spr.ColorCode)
			}
			disp.DrawSprite(pos.X, pos.Y, spr.SheetX, spr.SheetY, hexColor)
		} else if hasGlyph {
			glyph := w.Glyphs[i]
			hexColor := uint32(0xFFFFFFFF)
			if glyph.ColorCode != "" {
				hexColor = display.MapANSIColor(glyph.ColorCode)
			}
			disp.DrawText(pos.X, pos.Y, glyph.Char, hexColor)
		}
	}
}
