package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(800, 600, "Font Test")
	defer rl.CloseWindow()

	// Loading specifically without characters array to test stb_truetype default parsing
	font := rl.LoadFont("assets/fonts/NotoEmoji-Regular.ttf")

	fontChars := []rune{'👷', '⚡', 'A', '?', 'B'}
	for _, r := range fontChars {
		idx := rl.GetGlyphIndex(font, int32(r))
		fmt.Printf("Rune: %c (U+%04X), GlyphIndex: %d\n", r, r, idx)
	}
}
