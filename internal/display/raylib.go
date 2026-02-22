package display

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

type RaylibDisplay struct {
	TileSize int32
	FontSize int32
}

func NewRaylibDisplay(tileSize, fontSize int32) *RaylibDisplay {
	return &RaylibDisplay{
		TileSize: tileSize,
		FontSize: fontSize,
	}
}

func (r *RaylibDisplay) Init(gridWidth, gridHeight int, title string) error {
	rl.InitWindow(int32(gridWidth)*r.TileSize, int32(gridHeight)*r.TileSize, title)
	rl.SetTargetFPS(60) // Changed to 60 as per handover doc
	return nil
}

func (r *RaylibDisplay) Close() {
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
	rl.ClearBackground(rl.GetColor(colorHex))
}

func (r *RaylibDisplay) DrawText(gridX, gridY int, text string, colorHex uint32) {
	pixelX := int32(gridX) * r.TileSize
	pixelY := int32(gridY) * r.TileSize

	// Adjust Y to perfectly align font within the tile if necessary.
	// We'll just draw directly.
	rl.DrawText(text, pixelX, pixelY, r.FontSize, rl.GetColor(colorHex))
}

func (r *RaylibDisplay) PollInput() []core.InputEvent {
	var events []core.InputEvent

	if rl.IsKeyPressed(rl.KeyW) {
		events = append(events, core.InputEvent{Key: 'w'})
	}
	if rl.IsKeyPressed(rl.KeyS) {
		events = append(events, core.InputEvent{Key: 's'})
	}
	if rl.IsKeyPressed(rl.KeyA) {
		events = append(events, core.InputEvent{Key: 'a'})
	}
	if rl.IsKeyPressed(rl.KeyD) {
		events = append(events, core.InputEvent{Key: 'd'})
	}
	if rl.IsKeyPressed(rl.KeyP) {
		events = append(events, core.InputEvent{Key: 'p'})
	}
	if rl.IsKeyPressed(rl.KeyQ) {
		events = append(events, core.InputEvent{Key: 'q'})
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		events = append(events, core.InputEvent{Code: core.KeyEsc})
	}

	return events
}

// MapANSIColor Helper method for Color Mapper
func MapANSIColor(ansiColor string) uint32 {
	colorMap := map[string]uint32{
		world.Red:     0xFF0000FF,
		world.Green:   0x00FF00FF,
		world.Yellow:  0xFFFF00FF,
		world.Blue:    0x0000FFFF,
		world.Magenta: 0xFF00FFFF,
		world.Cyan:    0x00FFFFFF,
		world.White:   0xFFFFFFFF,
		world.Gray:    0x808080FF,
		"\x1b[97m":    0xFFFFFFFF, // Bright white
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
