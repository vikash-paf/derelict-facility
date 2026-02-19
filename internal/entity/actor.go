package entity

type Actor struct {
	X, Y int
	Char rune
}

func NewActor(x, y int, char rune) *Actor {
	return &Actor{
		X:    x,
		Y:    y,
		Char: char,
	}
}
