package display

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/vikash-paf/derelict-facility/internal/core"
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
	rl.ClearWindowState(rl.FlagWindowTransparent) // Fix transparency issue on some Linux window managers
	rl.InitWindow(int32(gridWidth)*r.CellWidth, int32(gridHeight)*r.CellHeight, title)
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)

	if r.FontPath != "" {
		var fontChars []rune
		for i := int32(32); i <= 126; i++ {
			fontChars = append(fontChars, rune(i))
		}
		extraChars := []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬', '█', '▓', '▒', '░', '·', '►', '◄', '▲', '▼', '⚡', '👷', '🖥'}
		fontChars = append(fontChars, extraChars...)

		r.Font = rl.LoadFontEx(r.FontPath, r.FontSize, fontChars)
		rl.SetTextureFilter(r.Font.Texture, rl.FilterPoint) // Pixel perfect text

		fallbackPath := "assets/fonts/NotoEmoji-Regular.ttf"
		r.FallbackFont = rl.LoadFontEx(fallbackPath, r.FontSize, fontChars)
	} else {
		r.Font = rl.GetFontDefault()
	}

	r.Tileset = rl.LoadTexture("assets/tileset.png")
	rl.SetTextureFilter(r.Tileset, rl.FilterPoint) // CRITICAL: Fixes gaps and blurring in pixel art

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

func (r *RaylibDisplay) Clear(color core.Color) {
	rl.ClearBackground(rl.Color{R: color.R, G: color.G, B: color.B, A: color.A})
}

func (r *RaylibDisplay) DrawRect(gridX, gridY int, color core.Color) {
	rl.DrawRectangle(
		int32(gridX)*r.CellWidth,
		int32(gridY)*r.CellHeight,
		r.CellWidth,
		r.CellHeight,
		rl.Color{R: color.R, G: color.G, B: color.B, A: color.A},
	)
}

func (r *RaylibDisplay) DrawText(gridX, gridY int, text string, color core.Color) {
	pixelY := int32(gridY) * r.CellHeight

	// Roger-style roguelikes use character-per-cell grid.
	// However, with large sprite cells, we want the text to be compact.
	// We will draw the entire string starting at the cell's top-left,
	// but we WON'T jump a full CellWidth per character.

	position := rl.NewVector2(float32(int32(gridX)*r.CellWidth), float32(pixelY))
	rl.DrawTextEx(r.Font, text, position, float32(r.FontSize), 1, rl.Color{R: color.R, G: color.G, B: color.B, A: color.A})
}

// DrawSprite cuts a frame out of the 4x4 Tileset atlas and draws it to the screen grid.
func (r *RaylibDisplay) DrawSprite(gridX, gridY int, sheetX, sheetY int, color core.Color) {
	// Dynamically calculate the tile size based on the texture dimensions (assuming 4x4 grid)
	spriteWidth := float32(r.Tileset.Width) / 4.0
	spriteHeight := float32(r.Tileset.Height) / 4.0

	// Where to cut the artwork on the giant sprite sheet
	sourceRec := rl.NewRectangle(float32(sheetX)*spriteWidth, float32(sheetY)*spriteHeight, spriteWidth, spriteHeight)

	// Where to draw the artwork on the game screen
	destRec := rl.NewRectangle(
		float32(int32(gridX)*r.CellWidth),
		float32(int32(gridY)*r.CellHeight),
		float32(r.CellWidth),
		float32(r.CellHeight),
	)

	origin := rl.NewVector2(0, 0)
	rl.DrawTexturePro(r.Tileset, sourceRec, destRec, origin, 0.0, rl.Color{R: color.R, G: color.G, B: color.B, A: color.A})
}

func (r *RaylibDisplay) PollInput() []core.InputEvent {
	var events []core.InputEvent

	if rl.IsKeyPressed(rl.KeyW) || rl.IsKeyPressedRepeat(rl.KeyW) {
		events = append(events, core.InputEvent{Key: rl.KeyW})
	}
	if rl.IsKeyPressed(rl.KeyS) || rl.IsKeyPressedRepeat(rl.KeyS) {
		events = append(events, core.InputEvent{Key: rl.KeyS})
	}
	if rl.IsKeyPressed(rl.KeyA) || rl.IsKeyPressedRepeat(rl.KeyA) {
		events = append(events, core.InputEvent{Key: rl.KeyA})
	}
	if rl.IsKeyPressed(rl.KeyD) || rl.IsKeyPressedRepeat(rl.KeyD) {
		events = append(events, core.InputEvent{Key: rl.KeyD})
	}
	if rl.IsKeyPressed(rl.KeyP) {
		events = append(events, core.InputEvent{Key: rl.KeyP})
	}
	if rl.IsKeyPressed(rl.KeyQ) {
		events = append(events, core.InputEvent{Key: rl.KeyQ})
	}
	if rl.IsKeyPressed(rl.KeyE) {
		events = append(events, core.InputEvent{Key: rl.KeyE})
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		events = append(events, core.InputEvent{Key: rl.KeyEscape})
	}
	return events
}

// DarkenColor takes a core.Color and reduces its brightness by the given factor.
func DarkenColor(color core.Color, factor uint8) core.Color {
	if factor <= 1 {
		return color
	}
	return core.Color{
		R: color.R / factor,
		G: color.G / factor,
		B: color.B / factor,
		A: color.A,
	}
}
