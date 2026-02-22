package components

import (
	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

// The Component Names as constants to prevent typos
const (
	NamePosition      = "Position"
	NameRenderable    = "Renderable"
	NamePlayerControl = "PlayerControl"
)

// Position holds exactly where an Entity is on the grid.
type Position struct {
	X int
	Y int
}

// Renderable determines what an Entity looks like when drawn to the screen.
type Renderable struct {
	Char      string
	ColorCode string // ANSI Color constant, e.g. "\x1b[31m" or world.Red
}

// PlayerControl indicates that this entity is currently controllable by the user.
// It also holds properties specific to their condition.
type PlayerControl struct {
	Autopilot   bool
	CurrentPath []entity.Point
	Status      world.PlayerStatus
}
