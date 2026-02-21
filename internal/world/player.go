package world

type PlayerStatus int

const (
	PlayerStatusHealthy PlayerStatus = iota
	PlayerStatusSick
	PlayerStatusHurt
)

const (
	playerCharacter = "🚶"
)

func NewPlayer(x, y int, status PlayerStatus) *Player {
	return &Player{x, y, status}
}

type Player struct {
	X, Y   int
	Status PlayerStatus
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
