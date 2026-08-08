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

// LerpColor smoothly interpolates between color c1 and c2 by factor t (0.0 to 1.0).
func LerpColor(c1, c2 Color, t float64) Color {
	if t <= 0 {
		return c1
	}
	if t >= 1 {
		return c2
	}
	return Color{
		R: uint8(float64(c1.R) + float64(int(c2.R)-int(c1.R))*t),
		G: uint8(float64(c1.G) + float64(int(c2.G)-int(c1.G))*t),
		B: uint8(float64(c1.B) + float64(int(c2.B)-int(c1.B))*t),
		A: uint8(float64(c1.A) + float64(int(c2.A)-int(c1.A))*t),
	}
}
