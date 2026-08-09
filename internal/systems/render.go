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
func RenderEntities(w *ecs.World, disp display.Display, gameMap *world.Map, bounds core.ViewportBounds) {
	renderEntity := func(i ecs.Entity) {
		// Must have a position to be rendered
		if (w.Masks[i] & components.MaskPosition) == 0 {
			return
		}

		hasSprite := (w.Masks[i] & components.MaskSprite) != 0
		hasGlyph := (w.Masks[i] & components.MaskGlyph) != 0

		// Must have at least one visual representation
		if !hasSprite && !hasGlyph {
			return
		}

		pos := w.Positions[i]

		// Culling: Only render if within the camera's view
		if pos.X < bounds.StartX || pos.X >= bounds.EndX || pos.Y < bounds.StartY || pos.Y >= bounds.EndY {
			return
		}

		// Is it the player? Check the mask for the PlayerControl bit
		isPlayer := w.IsPlayer(i)

		// Check if it's an active generator
		isGenerator := (w.Masks[i] & components.MaskPowerGenerator) != 0
		isActiveGenerator := isGenerator && w.PowerGenerators[i].IsActive

		if !isPlayer && !isActiveGenerator {
			tile := gameMap.GetTile(pos.X, pos.Y)
			if tile == nil || !tile.Visible {
				return
			}
		}

		if hasSprite {
			spr := w.Sprites[i]
			disp.DrawSprite(pos.X, pos.Y, spr.SheetX, spr.SheetY, spr.Color)
		} else if hasGlyph {
			glyph := w.Glyphs[i]
			disp.DrawText(pos.X, pos.Y, glyph.Char, glyph.Color)
		}
	}

	// Pass 1: Render all non-player entities
	for i := range ecs.Entity(ecs.MaxEntities) {
		isPlayer := w.IsPlayer(i)
		if !isPlayer {
			renderEntity(i)
		}
	}

	// Pass 2: Render all player entities
	for i := range ecs.Entity(ecs.MaxEntities) {
		isPlayer := w.IsPlayer(i)
		if isPlayer {
			renderEntity(i)
		}
	}
}
