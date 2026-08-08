package menu

import (
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/mission"
)

// MissionItem represents an entry in the Start Menu story selector.
type MissionItem struct {
	ID          string
	Title       string
	Description string
	Manifest    *mission.MissionManifest // nil for procedural / sandbox mode
}

// MenuState manages keyboard navigation and rendering for the game's start menu.
type MenuState struct {
	SelectedIndex int
	Items         []MissionItem
}

func NewMenuState() *MenuState {
	loader := mission.NewMissionLoader()
	discovered, _ := loader.DiscoverMissions()

	var items []MissionItem
	for _, m := range discovered {
		items = append(items, MissionItem{
			ID:          m.ID,
			Title:       m.Title,
			Description: m.Synopsis,
			Manifest:    &m,
		})
	}

	// Always provide infinite procedural sandbox mode as a menu option
	items = append(items, MissionItem{
		ID:          "procedural_sandbox",
		Title:       "PROCEDURAL DERELICT RUINS",
		Description: "Explore an infinite procedurally carved facility layout.",
		Manifest:    nil,
	})

	return &MenuState{
		SelectedIndex: 0,
		Items:         items,
	}
}

func (m *MenuState) SelectNext() {
	if len(m.Items) == 0 {
		return
	}
	m.SelectedIndex = (m.SelectedIndex + 1) % len(m.Items)
}

func (m *MenuState) SelectPrevious() {
	if len(m.Items) == 0 {
		return
	}
	m.SelectedIndex = (m.SelectedIndex - 1 + len(m.Items)) % len(m.Items)
}

func (m *MenuState) GetSelectedItem() MissionItem {
	if m.SelectedIndex < 0 || m.SelectedIndex >= len(m.Items) {
		return m.Items[0]
	}
	return m.Items[m.SelectedIndex]
}

// Render draws the sci-fi retro Main Menu on the display.
func (m *MenuState) Render(disp display.Display, viewWidth, viewHeight int) {
	disp.Clear(core.Color{R: 12, G: 15, B: 25, A: 255})

	centerX := viewWidth / 2

	// Render Title Header
	drawCenteredText(disp, centerX, 4, "=== DERELICT FACILITY ===", core.Cyan)
	drawCenteredText(disp, centerX, 5, "STORY CAMPAIGN & MISSION SELECTOR", core.Yellow)

	// Render Mission Items List
	startY := 9
	for i, item := range m.Items {
		lineY := startY + (i * 3)
		color := core.Gray
		prefix := "   "

		if i == m.SelectedIndex {
			color = core.BrightWhite
			prefix = "► "
		}

		drawCenteredText(disp, centerX, lineY, prefix+item.Title, color)
		drawCenteredText(disp, centerX, lineY+1, item.Description, core.DarkGray)
	}

	// Render Controls Footer
	footerY := viewHeight - 4
	drawCenteredText(disp, centerX, footerY, "[UP / DOWN] Select Campaign    [ENTER] Launch Mission    [Q] Quit", core.Green)
}

func drawCenteredText(disp display.Display, centerX, y int, text string, color core.Color) {
	x := centerX - (len(text) / 2)
	disp.DrawText(x, y, text, color)
}
