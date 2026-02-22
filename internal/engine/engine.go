package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

const (
	cursorHome = "\033[H"
	lineBreak  = "\r\n"
	fovRadius  = 8 // cool stuff can be done here, like a dimming torch light
)

type GameState uint8

const (
	GameStatePaused GameState = iota
	GameStateRunning
)

type Engine struct {
	Display    display.Display
	Map        *world.Map
	Player     *world.Player
	Theme      world.TileVariant
	TickerRate time.Duration
	tickCount  int
	State      GameState
	Running    bool
}

func NewEngine(
	disp display.Display,
	gameMap *world.Map,
	player *world.Player,
	startingTheme world.TileVariant,
) *Engine {
	e := &Engine{
		Display:    disp,
		Map:        gameMap,
		Player:     player,
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
		for _, event := range events {
			e.handleInput(event)
		}

		if e.State == GameStateRunning {
			e.Update() // Calculate all game rules!
		}

		e.render() // Paint the results!
	}

	return nil
}

func (e *Engine) handleInput(event core.InputEvent) {
	dx, dy := 0, 0

	if event.Quit {
		e.Running = false
		return
	}

	switch event.Key {
	case 'w':
		dy = -1
	case 's':
		dy = 1
	case 'a':
		dx = -1
	case 'd':
		dx = 1
	case 'p':
		e.Player.Autopilot = !e.Player.Autopilot
		e.Player.CurrentPath = nil // Clear the path if we toggle it
	case 'q':
		e.Running = false
		return
	case 27:
		// handle the escape button to toggle the game state
		if e.State == GameStateRunning {
			e.State = GameStatePaused
		} else {
			e.State = GameStateRunning
		}
	}

	if e.State == GameStateRunning && dx != 0 || dy != 0 {
		e.movePlayer(dx, dy)
	}
}

func (e *Engine) movePlayer(dx, dy int) {
	newX := e.Player.X + dx
	newY := e.Player.Y + dy

	// if outside the map, do nothing
	if newX < 0 || newX >= e.Map.Width || newY < 0 || newY >= e.Map.Height {
		return
	}

	// check if the tile is walkable
	tile := e.Map.GetTile(newX, newY)
	if tile != nil && tile.Walkable {
		e.Player.X = newX
		e.Player.Y = newY
	}
}

func (e *Engine) Update() {
	e.tickCount++

	switch e.State {
	case GameStatePaused:
		// do nothing, the world is frozen
		// later: implement it to save the game
	case GameStateRunning:
		e.processSimulation()
	}
}

func (e *Engine) processAutopilot() {
	// 1. If we don't have a path, find a new destination!
	if len(e.Player.CurrentPath) == 0 {
		// Pick a random room
		targetRoom := e.Map.Rooms[rand.Intn(len(e.Map.Rooms))]
		targetX, targetY := targetRoom.Center()

		start := entity.Point{X: e.Player.X, Y: e.Player.Y}
		target := entity.Point{X: targetX, Y: targetY}

		// Calculate the path
		path := world.FindPath(e.Map, start, target)

		// A* returns the starting node as index 0. We slice it off so we don't stand still.
		if len(path) > 1 {
			e.Player.CurrentPath = path[1:]
		} else {
			e.Player.CurrentPath = nil // We are already there, or trapped
		}

		return
	}

	// 2. Take the next step in the path
	nextStep := e.Player.CurrentPath[0]

	// Ensure the tile is still walkable (in case a door closed or enemy moved there)
	if e.Map.IsWalkable(nextStep.X, nextStep.Y) {
		e.Player.X = nextStep.X
		e.Player.Y = nextStep.Y
	} else {
		// Path is blocked! Clear it so we recalculate next tick.
		e.Player.CurrentPath = nil
		return
	}

	// 3. Pop the step we just took off the slice
	e.Player.CurrentPath = e.Player.CurrentPath[1:]
}

func (e *Engine) processSimulation() {
	// todo: this is where all the frame logic exists

	// Run AI movement every 2nd frame (approx 15 times a second)
	if e.Player.Autopilot && e.tickCount%2 == 0 {
		// run autopilot
		e.processAutopilot()
	}

	e.Map.ComputeFOV(e.Player.X, e.Player.Y, fovRadius)

	// this is where I will add:
	// - Check if the player stepped on a checkpoint
	// - Move enemies
	// - Trigger story events
	// - Update flashing light animations
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
	if e.Player.Autopilot {
		for _, p := range e.Player.CurrentPath {
			pathLookup[p.Y*e.Map.Width+p.X] = true
		}
	}

	for y := 0; y < e.Map.Height; y++ {
		for x := 0; x < e.Map.Width; x++ {
			// 1. Render the player
			if e.Player.X == x && e.Player.Y == y {
				text, color := display.ExtractTextAndColor(e.Player.Render())
				e.Display.DrawText(x, y, text, color)
				continue
			}

			tile := e.Map.GetTile(x, y)
			if tile == nil {
				continue
			}
			isPathTile := pathLookup[y*e.Map.Width+x]
			// We only draw the path if it's on a tile we've at least explored!
			// (Drawing a path through Pitch Black space breaks the Fog of War illusion).
			if isPathTile && (tile.Visible || tile.Explored) {
				e.Display.DrawText(x, y, "*", display.MapANSIColor(world.Red))
				continue
			}

			// 2. Render the map tiles
			if tile.Visible {
				text, color := display.ExtractTextAndColor(theme[tile.Type])
				e.Display.DrawText(x, y, text, color)
				continue
			}

			if tile.Explored {
				text, _ := display.ExtractTextAndColor(theme[tile.Type])
				e.Display.DrawText(x, y, text, display.MapANSIColor(world.Gray))
				continue
			}

			text, color := display.ExtractTextAndColor(theme[world.TileTypeEmpty])
			e.Display.DrawText(x, y, text, color)
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
	if e.Player.Status == world.PlayerStatusSick {
		statusText = "SICK / TOXIC"
	} else if e.Player.Status == world.PlayerStatusHurt {
		statusText = "CRITICAL"
	}
	e.drawText(2, hudY+1, fmt.Sprintf(" STATUS: %s ", statusText), world.Cyan)

	if e.Player.Autopilot {
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
