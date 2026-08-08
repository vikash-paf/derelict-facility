package menu

import (
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/display"
)

// MissionDef describes a story mission or campaign map.
type MissionDef struct {
	ID          string
	Title       string
	Description string
	MapFile     string // empty string means procedural generation
}

// DefaultMissions provides the list of available story missions and sandbox modes.
var DefaultMissions = []MissionDef{
	{
		ID:          "sector_4",
		Title:       "MISSION 1: SECTOR 4 RESEARCH FACILITY",
		Description: "Investigate containment breach in the sunlit glass skylight sector.",
		MapFile:     "test_map.json",
	},
	{
		ID:          "procedural",
		Title:       "MISSION 2: PROCEDURAL DERELICT RUINS",
		Description: "Explore an infinite procedurally carved derelict facility layout.",
		MapFile:     "",
	},
}

// MenuState manages keyboard navigation and rendering for the game's start menu.
type MenuState struct {
	SelectedIndex int
	Missions      []MissionDef
}

func NewMenuState() *MenuState {
	return &MenuState{
		SelectedIndex: 0,
		Missions:      DefaultMissions,
	}
}

func (m *MenuState) SelectNext() {
	if len(m.Missions) == 0 {
		return
	}
	m.SelectedIndex = (m.SelectedIndex + 1) % len(m.Missions)
}

func (m *MenuState) SelectPrevious() {
	if len(m.Missions) == 0 {
		return
	}
	m.SelectedIndex = (m.SelectedIndex - 1 + len(m.Missions)) % len(m.Missions)
}

func (m *MenuState) GetSelectedMission() MissionDef {
	if m.SelectedIndex < 0 || m.SelectedIndex >= len(m.Missions) {
		return DefaultMissions[0]
	}
	return m.Missions[m.SelectedIndex]
}

// Render draws the sci-fi retro Main Menu on the display.
func (m *MenuState) Render(disp display.Display, viewWidth, viewHeight int) {
	disp.Clear(core.Color{R: 12, G: 15, B: 25, A: 255})

	centerX := viewWidth / 2

	// Render Title Header
	drawCenteredText(disp, centerX, 4, "=== DERELICT FACILITY ===", core.Cyan)
	drawCenteredText(disp, centerX, 5, "STORY & CAMPAIGN SELECTOR", core.Yellow)

	// Render Missions List
	startY := 9
	for i, mission := range m.Missions {
		lineY := startY + (i * 3)
		color := core.Gray
		prefix := "   "

		if i == m.SelectedIndex {
			color = core.BrightWhite
			prefix = "► "
		}

		drawCenteredText(disp, centerX, lineY, prefix+mission.Title, color)
		drawCenteredText(disp, centerX, lineY+1, mission.Description, core.DarkGray)
	}

	// Render Controls Footer
	footerY := viewHeight - 4
	drawCenteredText(disp, centerX, footerY, "[UP / DOWN] Select Mission    [ENTER] Launch Campaign    [Q] Quit", core.Green)
}

func drawCenteredText(disp display.Display, centerX, y int, text string, color core.Color) {
	x := centerX - (len(text) / 2)
	disp.DrawText(x, y, text, color)
}
