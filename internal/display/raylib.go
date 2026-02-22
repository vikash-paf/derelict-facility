package display

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

type RaylibDisplay struct {
	CellWidth    int32
	CellHeight   int32
	FontSize     int32
	FontPath     string
	Font         rl.Font
	FallbackFont rl.Font
	Tileset      rl.Texture2D
}

func NewRaylibDisplay(cellWidth, cellHeight, fontSize int32, fontPath string) *RaylibDisplay {
	return &RaylibDisplay{
		CellWidth:  cellWidth,
		CellHeight: cellHeight,
		FontSize:   fontSize,
		FontPath:   fontPath,
	}
}

func (r *RaylibDisplay) Init(gridWidth, gridHeight int, title string) error {
	rl.InitWindow(int32(gridWidth)*r.CellWidth, int32(gridHeight)*r.CellHeight, title)
	rl.SetTargetFPS(30)
	rl.SetExitKey(0) // Disable the default Escape key exit behavior

	if r.FontPath != "" {
		// Define the characters we want to load from the font
		var fontChars []rune
		for i := int32(32); i <= 126; i++ {
			fontChars = append(fontChars, rune(i))
		}
		// Add some extended box drawing and roguelike characters, plus the emojis
		extraChars := []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬', '█', '▓', '▒', '░', '·', '►', '◄', '▲', '▼', '⚡', '👷'}
		fontChars = append(fontChars, extraChars...)

		r.Font = rl.LoadFontEx(r.FontPath, r.FontSize, fontChars)
		rl.SetTextureFilter(r.Font.Texture, rl.FilterBilinear)

		// Load the Noto Emoji font for fallback
		fallbackPath := "assets/fonts/NotoEmoji-Regular.ttf"
		// Raylib doesn't have a direct SetFontFallback in its current Go bindings,
		// but we can manage drawing fallbacks ourselves if needed, or rely on OS defaults.
		// Wait, rl-go does have it:
		r.FallbackFont = rl.LoadFontEx(fallbackPath, r.FontSize, fontChars)
	} else {
		r.Font = rl.GetFontDefault()
	}

	// Load the brand new beautiful Sci-Fi Graphics!
	r.Tileset = rl.LoadTexture("assets/kenney_sci-fi-rts/Tilesheet/scifi_tilesheet.png")

	return nil
}

func (r *RaylibDisplay) Close() {
	if r.FontPath != "" {
		rl.UnloadFont(r.Font)
		if r.FallbackFont.Texture.ID != 0 {
			rl.UnloadFont(r.FallbackFont)
		}
	}
	if r.Tileset.ID != 0 {
		rl.UnloadTexture(r.Tileset)
	}
	rl.CloseWindow()
}

func (r *RaylibDisplay) ShouldClose() bool {
	return rl.WindowShouldClose()
}

func (r *RaylibDisplay) BeginFrame() {
	rl.BeginDrawing()
}

func (r *RaylibDisplay) EndFrame() {
	rl.EndDrawing()
}

func (r *RaylibDisplay) Clear(colorHex uint32) {
	rl.ClearBackground(rl.GetColor(uint(colorHex)))
}

func (r *RaylibDisplay) DrawText(gridX, gridY int, text string, colorHex uint32) {
	pixelY := int32(gridY) * r.CellHeight

	colOffset := 0
	for _, char := range text {
		charStr := string(char)

		vecSize := rl.MeasureTextEx(r.Font, charStr, float32(r.FontSize), 0)
		charWidth := int32(vecSize.X)

		// Center character horizontally in the cell
		pixelX := int32(gridX+colOffset)*r.CellWidth + (r.CellWidth-charWidth)/2
		position := rl.NewVector2(float32(pixelX), float32(pixelY))

		// If the main font didn't load a glyph for this character, it returns an index of 0 (the '?')
		// We can check rl.GetGlyphIndex(r.Font, char) to see if it exists.
		glyphIndex := rl.GetGlyphIndex(r.Font, char)
		if glyphIndex == 0 && r.FallbackFont.Texture.ID != 0 && rl.GetGlyphIndex(r.FallbackFont, char) != 0 {
			rl.DrawTextEx(r.FallbackFont, charStr, position, float32(r.FontSize), 0, rl.GetColor(uint(colorHex)))
		} else {
			rl.DrawTextEx(r.Font, charStr, position, float32(r.FontSize), 0, rl.GetColor(uint(colorHex)))
		}

		colOffset++
	}
}

// DrawSprite cuts a 64x64 frame out of the Tileset atlas and draws it to the screen grid.
func (r *RaylibDisplay) DrawSprite(gridX, gridY int, sheetX, sheetY int, colorHex uint32) {
	// The kenney sprites are 64x64 grids
	spriteSize := float32(64)

	// Where to cut the artwork on the giant sprite sheet
	sourceRec := rl.NewRectangle(float32(sheetX)*spriteSize, float32(sheetY)*spriteSize, spriteSize, spriteSize)

	// Where to draw the artwork on the game screen
	// Note: We might need to stretch/squash it to fit our current grid cells, or draw it native 64x64!
	// Let's stretch it to fit our `r.CellWidth` and `r.CellHeight` exactly for now
	destRec := rl.NewRectangle(
		float32(int32(gridX)*r.CellWidth),
		float32(int32(gridY)*r.CellHeight),
		float32(r.CellWidth),
		float32(r.CellHeight),
	)

	// Pivot point is top-left
	origin := rl.NewVector2(0, 0)

	rl.DrawTexturePro(r.Tileset, sourceRec, destRec, origin, 0.0, rl.GetColor(uint(colorHex)))
}

func (r *RaylibDisplay) PollInput() []core.InputEvent {
	var events []core.InputEvent

	if rl.IsKeyPressed(rl.KeyW) || rl.IsKeyPressedRepeat(rl.KeyW) {
		events = append(events, core.InputEvent{Key: core.KeyW})
	}
	if rl.IsKeyPressed(rl.KeyS) || rl.IsKeyPressedRepeat(rl.KeyS) {
		events = append(events, core.InputEvent{Key: core.KeyS})
	}
	if rl.IsKeyPressed(rl.KeyA) || rl.IsKeyPressedRepeat(rl.KeyA) {
		events = append(events, core.InputEvent{Key: core.KeyA})
	}
	if rl.IsKeyPressed(rl.KeyD) || rl.IsKeyPressedRepeat(rl.KeyD) {
		events = append(events, core.InputEvent{Key: core.KeyD})
	}
	if rl.IsKeyPressed(rl.KeyP) {
		events = append(events, core.InputEvent{Key: core.KeyP})
	}
	if rl.IsKeyPressed(rl.KeyQ) {
		events = append(events, core.InputEvent{Key: core.KeyQ})
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		events = append(events, core.InputEvent{Key: core.KeyEsc})
	}

	return events
}

// MapANSIColor Helper method for Color Mapper
func MapANSIColor(ansiColor string) uint32 {
	colorMap := map[string]uint32{
		world.Red:         0xFF0000FF,
		world.Green:       0x00FF00FF,
		world.Yellow:      0xFFFF00FF,
		world.Blue:        0x0000FFFF,
		world.Magenta:     0xFF00FFFF,
		world.Cyan:        0x00FFFFFF,
		world.White:       0xFFFFFFFF,
		world.Gray:        0x808080FF,
		world.BrightWhite: 0xFFFFFFFF,
	}

	if hex, ok := colorMap[ansiColor]; ok {
		return hex
	}
	return 0xFFFFFFFF // default white
}

// ExtractTextAndColor takes a string like "\x1b[31m*\x1b[0m" and returns ("*", 0xFF0000FF).
// If it's a plain string like " ", it returns (" ", 0xFFFFFFFF)
func ExtractTextAndColor(s string) (string, uint32) {
	// Simple parsing since the patterns are very consistent: COLOR + CHAR + RESET
	// Or just plain CHAR
	if !strings.HasPrefix(s, "\x1b[") && !strings.HasPrefix(s, "\033[") {
		return s, 0xFFFFFFFF // Plain text, default white
	}

	// We assume color code ends with 'm' and reset is at the end.
	// Find first 'm'
	mIdx := strings.Index(s, "m")
	if mIdx == -1 {
		return s, 0xFFFFFFFF
	}

	colorCode := s[:mIdx+1]
	rest := s[mIdx+1:]

	// Strip the Reset code at the end
	text := strings.ReplaceAll(rest, world.Reset, "")

	return text, MapANSIColor(colorCode)
}
