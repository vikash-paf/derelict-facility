package display

import (
	"path"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/vikash-paf/derelict-facility/assets"
	"github.com/vikash-paf/derelict-facility/internal/core"
)

type RaylibDisplay struct {
	BaseCellWidth  int32
	BaseCellHeight int32
	BaseFontSize   int32
	CellWidth      int32
	CellHeight     int32
	FontSize       int32
	Scale          float32
	FontPath       string
	Font           rl.Font
	FallbackFont   rl.Font
	Tileset        rl.Texture2D
}

func NewRaylibDisplay(cellWidth, cellHeight, fontSize int32, fontPath string) *RaylibDisplay {
	return &RaylibDisplay{
		BaseCellWidth:  cellWidth,
		BaseCellHeight: cellHeight,
		BaseFontSize:   fontSize,
		CellWidth:      cellWidth,
		CellHeight:     cellHeight,
		FontSize:       fontSize,
		Scale:          1.0,
		FontPath:       fontPath,
	}
}

func (r *RaylibDisplay) SetScale(scale float32) {
	if scale < 0.5 {
		scale = 0.5
	}
	if scale > 3.0 {
		scale = 3.0
	}
	r.Scale = scale
	r.CellWidth = int32(float32(r.BaseCellWidth) * scale)
	r.CellHeight = int32(float32(r.BaseCellHeight) * scale)
	r.FontSize = int32(float32(r.BaseFontSize) * scale)
}

func (r *RaylibDisplay) GetScale() float32 {
	if r.Scale <= 0 {
		r.Scale = 1.0
	}
	return r.Scale
}

func (r *RaylibDisplay) Init(gridWidth, gridHeight int, title string) error {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(int32(gridWidth)*r.CellWidth, int32(gridHeight)*r.CellHeight, title)
	rl.ClearWindowState(rl.FlagWindowTransparent) // Fix transparency issue on some Linux window managers
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)

	if r.FontPath != "" {
		var fontChars []rune
		for i := int32(32); i <= 126; i++ {
			fontChars = append(fontChars, rune(i))
		}
		extraChars := []rune{'═', '║', '╔', '╗', '╚', '╝', '╠', '╣', '╦', '╩', '╬', '█', '▓', '▒', '░', '·', '►', '◄', '▲', '▼', '⚡', '👷', '🖥'}
		fontChars = append(fontChars, extraChars...)

		// Read font from embedded assets filesystem
		fontBytes, err := assets.AssetsFS.ReadFile(r.FontPath)
		if err == nil && len(fontBytes) > 0 {
			fileType := path.Ext(r.FontPath)
			r.Font = rl.LoadFontFromMemory(fileType, fontBytes, r.FontSize, fontChars)
		}

		if r.Font.Texture.ID == 0 || r.Font.BaseSize == 0 {
			// Fall back to default raylib font if loading fails
			r.Font = rl.GetFontDefault()
			r.FontPath = ""
		} else {
			rl.SetTextureFilter(r.Font.Texture, rl.FilterPoint) // Pixel perfect text

			fallbackPath := "fonts/NotoEmoji-Regular.ttf"
			fallbackBytes, err := assets.AssetsFS.ReadFile(fallbackPath)
			if err == nil && len(fallbackBytes) > 0 {
				r.FallbackFont = rl.LoadFontFromMemory(".ttf", fallbackBytes, r.FontSize, fontChars)
			}
		}
	} else {
		r.Font = rl.GetFontDefault()
	}

	// Read tileset texture from embedded assets if available
	tilesetBytes, err := assets.AssetsFS.ReadFile("tileset.png")
	if err == nil && len(tilesetBytes) > 0 {
		img := rl.LoadImageFromMemory(".png", tilesetBytes, int32(len(tilesetBytes)))
		r.Tileset = rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
		rl.SetTextureFilter(r.Tileset, rl.FilterPoint)
	}

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
	if rl.IsKeyPressed(rl.KeyEqual) || rl.IsKeyPressed(rl.KeyKpAdd) { // '+' or '='
		events = append(events, core.InputEvent{Key: rl.KeyEqual})
	}
	if rl.IsKeyPressed(rl.KeyMinus) || rl.IsKeyPressed(rl.KeyKpSubtract) { // '-'
		events = append(events, core.InputEvent{Key: rl.KeyMinus})
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

func (r *RaylibDisplay) GetDimensions() (int, int) {
	w := int(rl.GetScreenWidth()) / int(r.CellWidth)
	h := int(rl.GetScreenHeight()) / int(r.CellHeight)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

func (r *RaylibDisplay) IsResized() bool {
	return rl.IsWindowResized()
}
