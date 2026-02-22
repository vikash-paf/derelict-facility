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
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)

	if r.FontPath != "" {
		var fontChars []rune
		for i := int32(32); i <= 126; i++ {
			fontChars = append(fontChars, rune(i))
		}
		extraChars := []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬', '█', '▓', '▒', '░', '·', '►', '◄', '▲', '▼', '⚡', '👷'}
		fontChars = append(fontChars, extraChars...)

		r.Font = rl.LoadFontEx(r.FontPath, r.FontSize, fontChars)
		rl.SetTextureFilter(r.Font.Texture, rl.FilterPoint) // Pixel perfect text

		fallbackPath := "assets/fonts/NotoEmoji-Regular.ttf"
		r.FallbackFont = rl.LoadFontEx(fallbackPath, r.FontSize, fontChars)
	} else {
		r.Font = rl.GetFontDefault()
	}

	r.Tileset = rl.LoadTexture("assets/derelict_spritesheet.png")
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

func (r *RaylibDisplay) Clear(colorHex uint32) {
	rl.ClearBackground(rl.GetColor(uint(colorHex)))
}

func (r *RaylibDisplay) DrawText(gridX, gridY int, text string, colorHex uint32) {
	pixelY := int32(gridY) * r.CellHeight

	// Roger-style roguelikes use character-per-cell grid.
	// However, with large sprite cells, we want the text to be compact.
	// We will draw the entire string starting at the cell's top-left,
	// but we WON'T jump a full CellWidth per character.

	position := rl.NewVector2(float32(int32(gridX)*r.CellWidth), float32(pixelY))
	rl.DrawTextEx(r.Font, text, position, float32(r.FontSize), 1, rl.GetColor(uint(colorHex)))
}

// DrawSprite cuts a 128x128 frame out of the Tileset atlas and draws it to the screen grid.
func (r *RaylibDisplay) DrawSprite(gridX, gridY int, sheetX, sheetY int, colorHex uint32) {
	// Diagnosis: The spritesheet is actually a 128x128 grid (22x12 tiles).
	spriteSize := float32(128)

	sourceRec := rl.NewRectangle(float32(sheetX)*spriteSize, float32(sheetY)*spriteSize, spriteSize, spriteSize)

	destRec := rl.NewRectangle(
		float32(int32(gridX)*r.CellWidth),
		float32(int32(gridY)*r.CellHeight),
		float32(r.CellWidth),
		float32(r.CellHeight),
	)

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

func MapANSIColor(ansiColor string) uint32 {
	colorMap := map[string]uint32{
		world.Red: 0xFF0000FF, world.Green: 0x00FF00FF, world.Yellow: 0xFFFF00FF,
		world.Blue: 0x0000FFFF, world.Magenta: 0xFF00FFFF, world.Cyan: 0x00FFFFFF,
		world.White: 0xFFFFFFFF, world.Gray: 0x808080FF, world.BrightWhite: 0xFFFFFFFF,
	}
	if hex, ok := colorMap[ansiColor]; ok {
		return hex
	}
	return 0xFFFFFFFF
}

func ExtractTextAndColor(s string) (string, uint32) {
	if !strings.HasPrefix(s, "\x1b[") && !strings.HasPrefix(s, "\033[") {
		return s, 0xFFFFFFFF
	}
	mIdx := strings.Index(s, "m")
	if mIdx == -1 {
		return s, 0xFFFFFFFF
	}
	colorCode := s[:mIdx+1]
	text := strings.ReplaceAll(s[mIdx+1:], world.Reset, "")
	return text, MapANSIColor(colorCode)
}
