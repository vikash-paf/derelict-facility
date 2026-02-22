package world

import (
	"github.com/vikash-paf/derelict-facility/internal/entity"
)

type PlayerStatus int

const (
	PlayerStatusHealthy PlayerStatus = iota
	PlayerStatusSick
	PlayerStatusHurt
)

const (
	playerCharacter = "@"
)

func NewPlayer(x int, y int, status PlayerStatus, autopilot bool) *Player {
	return &Player{X: x, Y: y, Status: status, Autopilot: autopilot}
}

type Player struct {
	X, Y        int
	Status      PlayerStatus
	Autopilot   bool
	CurrentPath []entity.Point
}

func (p *Player) Render() string {
	switch p.Status {
	case PlayerStatusSick:
		// A sickly, toxic green
		return Green + playerCharacter + Reset
	case PlayerStatusHurt:
		// A flashing red warning
		return Red + playerCharacter + Reset
	default:
		// A bright, healthy white or cyan so they pop against the background
		return Cyan + playerCharacter + Reset
	}
}
