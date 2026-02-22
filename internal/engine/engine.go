package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/systems"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

const (
	fovRadius = 8 // cool stuff can be done here, like a dimming torch light
)

type GameState uint8

const (
	GameStatePaused GameState = iota
	GameStateRunning
)

type Engine struct {
	Display    display.Display
	Map        *world.Map
	EcsWorld   *ecs.World // Replaces Player
	Theme      world.TileVariant
	TickerRate time.Duration
	tickCount  int
	State      GameState
	Running    bool
}

func NewEngine(
	disp display.Display,
	gameMap *world.Map,
	ecsWorld *ecs.World,
	startingTheme world.TileVariant,
) *Engine {
	e := &Engine{
		Display:    disp,
		Map:        gameMap,
		EcsWorld:   ecsWorld,
		State:      GameStateRunning,
		Running:    true,
		TickerRate: time.Millisecond * 33, // ~30 fps
		Theme:      startingTheme,
	}

	return e
}

// Run starts the deterministic game loop
func (e *Engine) Run() error {
	for !e.Display.ShouldClose() && e.Running {
		events := e.Display.PollInput()
		e.handleInputForGlobals(events)

		if e.State == GameStateRunning {
			e.Update(events) // Calculate all game rules!
		}

		e.render() // Paint the results!
	}

	return nil
}

func (e *Engine) handleInputForGlobals(events []core.InputEvent) {
	for _, event := range events {
		if event.Quit || event.Key == core.KeyQ {
			e.Running = false
			return
		}
		if event.Key == core.KeyEsc { // Escape
			if e.State == GameStateRunning {
				e.State = GameStatePaused
			} else {
				e.State = GameStateRunning
			}
		}
	}
}

func (e *Engine) Update(events []core.InputEvent) {
	e.tickCount++

	switch e.State {
	case GameStatePaused:
		// do nothing, the world is frozen
		// later: implement it to save the game
	case GameStateRunning:
		e.processSimulation(events)
	}
}

func (e *Engine) processAutopilot() {
	// Replaced by systems.ProcessAutopilot
}

func (e *Engine) processSimulation(events []core.InputEvent) {
	// Let the systems tick using the events we polled at the start of the frame!
	systems.ProcessPlayerInput(e.EcsWorld, events, e.Map)

	// Run AI movement every 2nd frame (approx 15 times a second)
	if e.tickCount%2 == 0 {
		systems.ProcessAutopilot(e.EcsWorld, e.Map)
	}

	// Calculate FOV (We need to find the player's position first in an ECS)
	playerEntities := e.EcsWorld.GetEntitiesWith(components.NamePlayerControl)
	if len(playerEntities) > 0 {
		if posRaw := e.EcsWorld.GetComponent(playerEntities[0], components.NamePosition); posRaw != nil {
			pos := posRaw.(*components.Position)
			e.Map.ComputeFOV(pos.X, pos.Y, fovRadius)
		}
	}
}

func (e *Engine) Pause() {
	e.State = GameStatePaused
}

func (e *Engine) Resume() {
	e.State = GameStateRunning
}

// render updates the game screen by drawing the map, GameState overlays,
// and other visual elements to the Display buffer.
//
// Renders top to bottom as separate layers.
func (e *Engine) render() {
	e.Display.BeginFrame()
	e.Display.Clear(0x000000FF) // Black background

	renderTheme := e.Theme

	if e.State == GameStatePaused {
		renderTheme = world.TileVariantPaused
	}

	e.renderMapLayer(renderTheme)
	systems.RenderEntities(e.EcsWorld, e.Display, e.Map)
	e.renderHUD()

	switch e.State {
	case GameStatePaused:
		e.renderPauseMenu()
	default:
	}

	e.Display.EndFrame()
}

func (e *Engine) renderPauseMenu() {
	e.drawTextCentered(14, "=== SYSTEM PAUSED ===", world.Red)
	e.drawTextCentered(16, "Press [ESC] to Resume", world.White)
	e.drawTextCentered(17, "Press [Q] to Quit", world.Gray)
}

func (e *Engine) renderMapLayer(theme world.TileVariant) {
	pathLookup := make(map[int]bool)

	// Collect paths from all PlayerControl entities to draw the red autopilot line
	playerEntities := e.EcsWorld.GetEntitiesWith(components.NamePlayerControl)
	for _, ent := range playerEntities {
		if ctrlRaw := e.EcsWorld.GetComponent(ent, components.NamePlayerControl); ctrlRaw != nil {
			ctrl := ctrlRaw.(*components.PlayerControl)
			if ctrl.Autopilot {
				for _, p := range ctrl.CurrentPath {
					pathLookup[p.Y*e.Map.Width+p.X] = true
				}
			}
		}
	}

	for y := 0; y < e.Map.Height; y++ {
		for x := 0; x < e.Map.Width; x++ {
			tile := e.Map.GetTile(x, y)
			if tile == nil {
				continue
			}
			isPathTile := pathLookup[y*e.Map.Width+x]
			// We only draw the path if it's on a tile we've at least explored!
			// (Drawing a path through Pitch Black space breaks the Fog of War illusion).
			if isPathTile && (tile.Visible || tile.Explored) {
				e.Display.DrawSprite(x, y, 22, 12, display.MapANSIColor(world.Red)) // '22, 12' as a placeholder path pip
				continue
			}

			// 2. Map TileType to a Sprite Coordinate
			// Since we don't know the exact Kenney Sci-Fi sheet coordinates, we will use educated guesses.
			// The user can easily change these later to point to the correct door, wall, and floor variations.
			sheetX, sheetY := 0, 0
			switch tile.Type {
			case world.TileTypeWall:
				sheetX, sheetY = 0, 1 // Assuming Row 1 Col 0 is a basic wall block
			case world.TileTypeFloor:
				sheetX, sheetY = 1, 0 // Assuming Row 0 Col 1 is a floor panel
			case world.TileTypeEmpty:
				sheetX, sheetY = 0, 0 // Empty transparent/black tile
			}

			if tile.Visible {
				_, color := display.ExtractTextAndColor(theme[tile.Type])
				e.Display.DrawSprite(x, y, sheetX, sheetY, color)
				continue
			}

			if tile.Explored {
				e.Display.DrawSprite(x, y, sheetX, sheetY, display.MapANSIColor(world.Gray))
				continue
			}

			// Unexplored/Empty Space
			e.Display.DrawSprite(x, y, 0, 0, 0x000000FF)
		}
	}
}

func (e *Engine) renderHUD() {
	// The Y-coordinate where the map ends and the HUD begins
	hudY := e.Map.Height

	// strings.Repeat is a highly optimized Go standard library function
	divider := strings.Repeat("═", e.Map.Width)
	e.drawText(0, hudY, divider, world.Gray)

	statusText := "HEALTHY"
	autopilotEngaged := false

	// Find player state for HUD
	playerEntities := e.EcsWorld.GetEntitiesWith(components.NamePlayerControl)
	if len(playerEntities) > 0 {
		if ctrlRaw := e.EcsWorld.GetComponent(playerEntities[0], components.NamePlayerControl); ctrlRaw != nil {
			ctrl := ctrlRaw.(*components.PlayerControl)
			autopilotEngaged = ctrl.Autopilot
			if ctrl.Status == components.PlayerStatusSick {
				statusText = "SICK / TOXIC"
			} else if ctrl.Status == components.PlayerStatusHurt {
				statusText = "CRITICAL"
			}
		}
	}

	e.drawText(2, hudY+1, fmt.Sprintf(" STATUS: %s ", statusText), world.Cyan)

	if autopilotEngaged {
		e.drawText(25, hudY+1, "[ NAV-COM: AUTOPILOT ENGAGED ]", world.Red)
	} else {
		e.drawText(25, hudY+1, "[ NAV-COM: MANUAL OVERRIDE ]  ", world.Gray)
	}

	// %06d formats the integer to always be 6 digits (e.g., 000142)
	cycleText := fmt.Sprintf(" CYCLE: %06d ", e.tickCount)
	e.drawText(e.Map.Width-len(cycleText)-2, hudY+1, cycleText, world.White)

	controls := " [W/A/S/D] Move    [P] Toggle Autopilot    [ESC] Pause System    [Q] Abort"
	e.drawText(2, hudY+2, controls, world.Gray)
}

func (e *Engine) drawTextCentered(y int, text string, colorCode string) {
	centerX := e.Map.Width / 2
	halfText := len(text) / 2
	x := centerX - halfText

	textStr, _ := display.ExtractTextAndColor(text)
	colorHex := display.MapANSIColor(colorCode)

	e.Display.DrawText(x, y, textStr, colorHex)
}

func (e *Engine) drawText(x, y int, text string, colorCode string) {
	textStr, _ := display.ExtractTextAndColor(text)
	colorHex := display.MapANSIColor(colorCode)

	e.Display.DrawText(x, y, textStr, colorHex)
}
