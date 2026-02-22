package systems

import (
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
)

// RenderEntities loops through all entities possessing BOTH a Renderable and Position component
// and draws them to the active display buffer.
func RenderEntities(w *ecs.World, disp display.Display) {
	// A more robust ECS would have a function to get entities with multiple components,
	// but for simplicity we iterate over one and check for the other.
	entities := w.GetEntitiesWith(components.NameRenderable)
	for _, e := range entities {
		renderRaw := w.GetComponent(e, components.NameRenderable)
		posRaw := w.GetComponent(e, components.NamePosition)

		if renderRaw == nil || posRaw == nil {
			continue // Need both to draw!
		}

		renderCfg := renderRaw.(*components.Renderable)
		pos := posRaw.(*components.Position)

		text, hexColor := display.ExtractTextAndColor(renderCfg.Char)
		// Usually the component would just hold a hex ID natively to avoid parsing,
		// but since your strings contain ANSI, we extract them here.
		if renderCfg.ColorCode != "" {
			hexColor = display.MapANSIColor(renderCfg.ColorCode)
		}

		disp.DrawText(pos.X, pos.Y, text, hexColor)
	}
}
