package core

type InputEvent struct {
	Key  rune
	Code int // for non-runes
	Quit bool
}
