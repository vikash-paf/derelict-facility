package display

import (
	"github.com/vikash-paf/derelict-facility/internal/core"
)

type Display interface {
	Init(gridWidth, gridHeight int, title string) error
	Close()
	ShouldClose() bool
	BeginFrame()
	EndFrame()
	Clear(color core.Color)
	DrawRect(gridX, gridY int, color core.Color)
	DrawText(gridX, gridY int, text string, color core.Color)
	DrawSprite(gridX, gridY int, sheetX, sheetY int, color core.Color)
	PollInput() []core.InputEvent
	GetDimensions() (int, int)
	IsResized() bool
	SetScale(scale float32)
	GetScale() float32
	ToggleShader()
}
