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
	// The mask we care about: Must have BOTH a Sprite AND a Position
	targetMask := components.MaskSprite | components.MaskPosition

	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		// Does this entity index have all the required bits turned on?
		if (w.Masks[i] & targetMask) == targetMask {

			// Notice how clean this is? No type casting!
			pos := w.Positions[i]
			spr := w.Sprites[i]

			// Is it the player? Check the mask for the PlayerControl bit
			isPlayer := (w.Masks[i] & components.MaskPlayerControl) != 0

			if !isPlayer {
				tile := gameMap.GetTile(pos.X, pos.Y)
				if tile == nil || !tile.Visible {
					continue
				}
			}

			hexColor := uint32(0xFFFFFFFF)
			if spr.ColorCode != "" {
				hexColor = display.MapANSIColor(spr.ColorCode)
			}

			disp.DrawSprite(pos.X, pos.Y, spr.SheetX, spr.SheetY, hexColor)
		}
	}
}
