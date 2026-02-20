package world

type PlayerStatus int

const (
	StatusHealthy PlayerStatus = iota
	StatusSick
	StatusHurt
)

type Player struct {
	X, Y   int
	Status PlayerStatus
}

func (p *Player) Render() string {
	switch p.Status {
	case StatusSick:
		// A sickly, toxic green
		return Green + "@" + Reset
	case StatusHurt:
		// A flashing red warning
		return Red + "@" + Reset
	default:
		// A bright, healthy white or cyan so they pop against the background
		return Cyan + "@" + Reset
	}
}
