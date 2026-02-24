package ecs

import "github.com/vikash-paf/derelict-facility/internal/components"

// Entity is just an index into the World arrays.
type Entity uint32

const MaxEntities = 1000

// World manages all entities and their component data in flat arrays (SoA).
type World struct {
	nextEntityID Entity
	freeEntities []Entity // Queue of IDs from destroyed entities we can reuse

	// The Component Mask array. If Masks[5] has the MaskPosition bit set,
	// it means Positions[5] contains valid data for Entity 5.
	Masks [MaxEntities]components.ComponentMask

	// The Dense Component Arrays (Structure of Arrays)
	Positions       [MaxEntities]components.Position
	Sprites         [MaxEntities]components.Sprite
	PlayerControls  [MaxEntities]components.PlayerControl
	Glyphs          [MaxEntities]components.Glyph
	Interactables   [MaxEntities]components.Interactable
	PowerGenerators [MaxEntities]components.PowerGenerator
}

func NewWorld() *World {
	return &World{
		nextEntityID: 0, // Start at 0 so it aligns with array indices!
		freeEntities: make([]Entity, 0),
	}
}

func (w *World) CreateEntity() Entity {
	var id Entity
	if len(w.freeEntities) > 0 {
		// Pop an ID off the free list
		id = w.freeEntities[len(w.freeEntities)-1]
		w.freeEntities = w.freeEntities[:len(w.freeEntities)-1]
	} else {
		id = w.nextEntityID
		w.nextEntityID++
		// If we exceed MaxEntities in a real game, we'd need to grow the arrays or panic.
		if id >= MaxEntities {
			panic("Max entities reached!")
		}
	}

	w.Masks[id] = components.MaskNone // Clear any old data mask
	return id
}

func (w *World) DestroyEntity(e Entity) {
	w.Masks[e] = components.MaskNone // Unset all bits. The data stays in RAM, but systems will ignore it.
	w.freeEntities = append(w.freeEntities, e)
}

func (w *World) AddPosition(e Entity, pos components.Position) {
	w.Positions[e] = pos
	w.Masks[e] |= components.MaskPosition // Turn ON the bit
}

func (w *World) AddSprite(e Entity, spr components.Sprite) {
	w.Sprites[e] = spr
	w.Masks[e] |= components.MaskSprite
}

func (w *World) AddPlayerControl(e Entity, ctrl components.PlayerControl) {
	w.PlayerControls[e] = ctrl
	w.Masks[e] |= components.MaskPlayerControl
}

func (w *World) AddGlyph(e Entity, glyph components.Glyph) {
	w.Glyphs[e] = glyph
	w.Masks[e] |= components.MaskGlyph
}

func (w *World) AddSolid(e Entity) {
	w.Masks[e] |= components.MaskSolid // No data to store, just set the flag!
}

func (w *World) AddInteractable(e Entity, interactable components.Interactable) {
	w.Interactables[e] = interactable
	w.Masks[e] |= components.MaskInteractable
}

func (w *World) AddPowerGenerator(e Entity, gen components.PowerGenerator) {
	w.PowerGenerators[e] = gen
	w.Masks[e] |= components.MaskPowerGenerator
}
