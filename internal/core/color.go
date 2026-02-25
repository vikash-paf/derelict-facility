package core

// Color represents an RGBA color value.
type Color struct {
	R, G, B, A uint8
}

// Predefined colors
var (
	Black       = Color{0, 0, 0, 255}
	White       = Color{255, 255, 255, 255}
	Red         = Color{255, 0, 0, 255}
	Green       = Color{0, 255, 0, 255}
	Blue        = Color{0, 0, 255, 255}
	Yellow      = Color{255, 255, 0, 255}
	Magenta     = Color{255, 0, 255, 255}
	Cyan        = Color{0, 255, 255, 255}
	Gray        = Color{128, 128, 128, 255}
	DarkGray    = Color{64, 64, 64, 255}
	BrightWhite = Color{255, 255, 255, 255}
)
