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
	Clear(colorHex uint32)
	DrawRect(gridX, gridY int, colorHex uint32)
	DrawText(gridX, gridY int, text string, colorHex uint32)
	DrawSprite(gridX, gridY int, sheetX, sheetY int, colorHex uint32)
	PollInput() []core.InputEvent
}
