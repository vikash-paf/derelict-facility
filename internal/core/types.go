package core

type InputEvent struct {
	Key  rune
	Code int // for non-runes
	Quit bool
}

// ViewportBounds defines the visible grid area.
type ViewportBounds struct {
	StartX int
	EndX   int
	StartY int
	EndY   int
}

