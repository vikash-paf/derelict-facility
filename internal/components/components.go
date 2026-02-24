package components

import (
	"github.com/vikash-paf/derelict-facility/internal/entity"
)

// ComponentMask is a bitmask representing the components an entity has.
type ComponentMask uint32

// Define a unique bit for every component type.
const (
	MaskNone          ComponentMask = 0
	MaskPosition      ComponentMask = 1 << iota // 1 (001)
	MaskSprite                                  // 2 (010)
	MaskPlayerControl                           // 4 (100)
)

// PlayerStatus represents the health/condition of a player entity.
type PlayerStatus uint8

const (
	PlayerStatusHealthy PlayerStatus = iota
	PlayerStatusSick
	PlayerStatusHurt
)

// Position holds exactly where an Entity is on the grid.
type Position struct {
	X int
	Y int
}

// Sprite defines the graphical representation of an entity using a 2D Texture Atlas.
// SheetX and SheetY represent the zero-indexed grid coordinates on the tileset image.
// e.g., SheetX: 0, SheetY: 0 is the top-left tile on the PNG.
type Sprite struct {
	SheetX    int
	SheetY    int
	ColorCode string // To optionally colorize or tint the sprite
}

// PlayerControl indicates that this entity is currently controllable by the user.
// It also holds properties specific to their condition.
type PlayerControl struct {
	Autopilot   bool
	CurrentPath []entity.Point
	Status      PlayerStatus
}
