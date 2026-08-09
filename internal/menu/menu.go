package menu

import (
	"strings"

	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/mission"
)

type MissionItem struct {
	ID          string
	Title       string
	Description string
	Manifest    *mission.MissionManifest
}

type MenuState struct {
	SelectedIndex int
	Items         []MissionItem
}

func NewMenuState() *MenuState {
	loader := mission.NewMissionLoader()
	discovered, _ := loader.DiscoverMissions()

	var items []MissionItem
	for _, m := range discovered {
		mCopy := new(mission.MissionManifest)
		*mCopy = m
		items = append(items, MissionItem{
			ID:          mCopy.ID,
			Title:       mCopy.Title,
			Description: mCopy.Synopsis,
			Manifest:    mCopy,
		})
	}

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

func (m *MenuState) Render(disp display.Display, viewWidth, viewHeight int) {
	disp.Clear(core.Color{R: 12, G: 15, B: 25, A: 255})

	centerX := viewWidth / 2

	drawCenteredText(disp, centerX, 4, "=== DERELICT FACILITY ===", core.Cyan)
	drawCenteredText(disp, centerX, 5, "STORY CAMPAIGN & MISSION SELECTOR", core.Yellow)

	// Render Mission Items List
	currentY := 9
	for i, item := range m.Items {
		color := core.Gray
		prefix := "   "

		if i == m.SelectedIndex {
			color = core.BrightWhite
			prefix = "► "
		}

		drawCenteredText(disp, centerX, currentY, prefix+item.Title, color)
		currentY++

		descLines := wrapText(item.Description, 75)
		for _, dLine := range descLines {
			drawCenteredText(disp, centerX, currentY, dLine, core.DarkGray)
			currentY++
		}

		currentY++
	}

	// Render Controls Footer
	footerY := viewHeight - 4
	drawCenteredText(disp, centerX, footerY, "[UP / DOWN] Select Campaign    [ENTER] Launch Mission    [Q] Quit", core.Green)
}

func drawCenteredText(disp display.Display, centerX, y int, text string, color core.Color) {
	x := centerX - (len(text) / 2)
	disp.DrawText(x, y, text, color)
}

func wrapText(text string, limit int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) > limit {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine += " " + word
		}
	}
	lines = append(lines, currentLine)
	return lines
}
