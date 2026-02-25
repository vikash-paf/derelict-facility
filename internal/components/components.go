package components

import (
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/entity"
)

// ComponentMask is a bitmask representing the components an entity has.
type ComponentMask uint32

// Define a unique bit for every component type.
const (
	MaskNone     ComponentMask = 0
	MaskPosition ComponentMask = 1 << iota
	MaskSprite
	MaskPlayerControl
	MaskGlyph
	MaskSolid
	MaskInteractable
	MaskPowerGenerator
	MaskDoor
	MaskTerminal
)

// PlayerStatus represents the health/condition of a player entity.
type PlayerStatus uint8

const (
	PlayerStatusHealthy PlayerStatus = iota
	PlayerStatusSick
	PlayerStatusHurt
)

func (s PlayerStatus) Title() string {
	switch s {
	case PlayerStatusHealthy:
		return "Healthy"
	case PlayerStatusSick:
		return "SICK / TOXIC"
	case PlayerStatusHurt:
		return "Hurt"
	default:
		return "Unknown"
	}
}

// Position holds exactly where an Entity is on the grid.
type Position struct {
	X int
	Y int
}

// Sprite defines the graphical representation of an entity using a 2D Texture Atlas.
// SheetX and SheetY represent the zero-indexed grid coordinates on the tileset image.
// e.g., SheetX: 0, SheetY: 0 is the top-left tile on the PNG.
type Sprite struct {
	SheetX int
	SheetY int
	Color  core.Color // To optionally colorize or tint the sprite
}

// PlayerControl indicates that this entity is currently controllable by the user.
// It also holds properties specific to their condition.
type PlayerControl struct {
	Autopilot   bool
	CurrentPath []entity.Point
	Status      PlayerStatus
}

// Glyph defines the graphical representation of an entity using a text character or emoji.
type Glyph struct {
	Char  string
	Color core.Color // RGBA color
}

// Solid indicates this entity cannot be walked through.
type Solid struct{} // empty struct because the bitmask itself holds the logic!

// Interactable allows the player to trigger an action when standing nearby and pressing [E].
type Interactable struct {
	Prompt string
}

// PowerGenerator is a specific interactive device state.
type PowerGenerator struct {
	IsActive bool
}

// Door represents a mechanism that can block movement and vision.
type Door struct {
	IsOpen bool
}

// Terminal allows saving the game.
type Terminal struct {
	HasSaved bool
}

