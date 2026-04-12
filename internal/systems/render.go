package systems

import (
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// RenderEntities loops through all entities possessing BOTH a Sprite or Glyph and Position component
// and draws them to the active display buffer if they are within exactly visible map tiles and camera viewport.
func RenderEntities(w *ecs.World, disp display.Display, gameMap *world.Map, cam *core.Camera) {
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

		// Culling: Only render if within the camera's view
		if !cam.IsInView(pos.X, pos.Y) {
			continue
		}

		// Is it the player? Check the mask for the PlayerControl bit
		isPlayer := (w.Masks[i] & components.MaskPlayerControl) != 0

		// Check if it's an active generator
		isGenerator := (w.Masks[i] & components.MaskPowerGenerator) != 0
		isActiveGenerator := isGenerator && w.PowerGenerators[i].IsActive

		if !isPlayer && !isActiveGenerator {
			tile := gameMap.GetTile(pos.X, pos.Y)
			if tile == nil || !tile.Visible {
				continue
			}
		}

		screenX, screenY := cam.WorldToScreen(pos.X, pos.Y)

		if hasSprite {
			spr := w.Sprites[i]
			disp.DrawSprite(screenX, screenY, spr.SheetX, spr.SheetY, spr.Color)
		} else if hasGlyph {
			glyph := w.Glyphs[i]
			disp.DrawText(screenX, screenY, glyph.Char, glyph.Color)
		}
	}
}
