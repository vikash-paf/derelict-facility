package engine

import (
	"fmt"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
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

// Flip the game state without branching, so it's a bit faster than using a switch statement.
func (s GameState) Flip() GameState {
	return s ^ 1
}

type Engine struct {
	Display    display.Display
	Map        *world.Map
	EcsWorld   *ecs.World // Replaces Player
	BaseTheme  world.TileVariant
	TickerRate time.Duration
	tickCount  int
	Clock      *world.FacilityClock
	State      GameState
	Running    bool
	PathLookup []bool // Pre-allocated array to avoid map allocations per frame
	Pathfinder *world.Pathfinder
	Messages   []string
	Camera     *core.Camera
}

func NewEngine(
	disp display.Display,
	gameMap *world.Map,
	ecsWorld *ecs.World,
	startingTheme world.TileVariant,
	viewWidth, viewHeight int,
) *Engine {
	e := &Engine{
		Display:    disp,
		Map:        gameMap,
		EcsWorld:   ecsWorld,
		State:      GameStateRunning,
		Running:    true,
		TickerRate: time.Millisecond * 33, // ~30 fps
		BaseTheme:  startingTheme,
		Clock:      world.NewFacilityClock(),
		PathLookup: make([]bool, gameMap.Width*gameMap.Height),
		Pathfinder: world.NewPathfinder(gameMap.Width, gameMap.Height),
		Camera: &core.Camera{
			X:      0,
			Y:      0,
			Width:  viewWidth,
			Height: viewHeight,
		},
	}

	return e
}

// Run starts the deterministic game loop
func (e *Engine) Run() error {
	for !e.Display.ShouldClose() && e.Running {
		if e.Display.IsResized() {
			e.recalculateViewport()
		}

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
		if event.Quit || event.Key == rl.KeyQ {
			e.Running = false
			return
		}
		if event.Key == rl.KeyEscape {
			e.State = e.State.Flip()
		}
		if event.Key == rl.KeyEqual { // '+' key to Zoom In
			currentScale := e.Display.GetScale()
			e.Display.SetScale(currentScale + 0.25)
			e.recalculateViewport()
		}
		if event.Key == rl.KeyMinus { // '-' key to Zoom Out
			currentScale := e.Display.GetScale()
			e.Display.SetScale(currentScale - 0.25)
			e.recalculateViewport()
		}
	}
}

func (e *Engine) recalculateViewport() {
	gw, gh := e.Display.GetDimensions()
	hudHeight := 8
	if gh > hudHeight {
		e.Camera.Width = gw
		e.Camera.Height = gh - hudHeight
	} else {
		e.Camera.Width = gw
		e.Camera.Height = gh
	}
}

func (e *Engine) Update(events []core.InputEvent) {
	e.tickCount++

	switch e.State {
	case GameStatePaused:
		// do nothing, the world is frozen
	case GameStateRunning:
		e.Clock.Tick()
		e.processSimulation(events)
	}
}

func (e *Engine) processSimulation(events []core.InputEvent) {
	// Let the systems tick using the events we polled at the start of the frame!
	systems.ProcessPlayerInput(e.EcsWorld, events, e.Map, func(msg string) {
		// If the new message is the same as the last one, don't repeat it
		if len(e.Messages) > 0 && e.Messages[len(e.Messages)-1] == msg {
			return
		}
		e.Messages = append(e.Messages, msg)
		// Limit to last 3 messages for UI space
		if len(e.Messages) > 3 {
			e.Messages = e.Messages[1:]
		}
	})

	// Center camera on player
	e.updateCamera()

	// Run AI movement every 2nd frame (approx 15 times a second)
	if e.tickCount%6 == 0 {
		systems.ProcessAutopilot(e.EcsWorld, e.Map, e.Pathfinder)
	}

	// Calculate FOV
	targetMask := components.MaskPlayerControl | components.MaskPosition
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			pos := e.EcsWorld.Positions[i]

			e.Map.ComputeFOV(pos.X, pos.Y, fovRadius, func(x, y int) bool {
				// 1. Is the map tile a wall?
				if !e.Map.IsWalkable(x, y) {
					return true
				}
				// 2. Is there a Solid entity (like a closed door)?
				return systems.IsSolidAt(e.EcsWorld, x, y)
			}, func(x, y int) bool {
				// Lit by power grid generator OR lit by natural daytime skylight
				isPowered := systems.IsPowerActiveAt(e.EcsWorld, e.Map, x, y)
				if isPowered {
					return true
				}
				tile := e.Map.GetTile(x, y)
				return tile != nil && tile.IsSunlit && e.Clock.IsDaytime()
			})
			break // Compute FOV for the first player found
		}
	}
}

func (e *Engine) Pause() {
	e.State = GameStatePaused
}

func (e *Engine) Resume() {
	e.State = GameStateRunning
}

func (e *Engine) updateCamera() {
	targetMask := components.MaskPlayerControl | components.MaskPosition
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			pos := e.EcsWorld.Positions[i]

			// Goal: player in the middle
			newX := pos.X - (e.Camera.Width / 2)
			newY := pos.Y - (e.Camera.Height / 2)

			// Clamp to map boundaries
			if newX < 0 {
				newX = 0
			}
			if newY < 0 {
				newY = 0
			}
			if newX > e.Map.Width-e.Camera.Width {
				newX = e.Map.Width - e.Camera.Width
			}
			if newY > e.Map.Height-e.Camera.Height {
				newY = e.Map.Height - e.Camera.Height
			}

			// Don't let the camera go negative if the map is smaller than the viewport
			if e.Map.Width < e.Camera.Width {
				newX = 0
			}
			if e.Map.Height < e.Camera.Height {
				newY = 0
			}

			e.Camera.X = newX
			e.Camera.Y = newY
			break
		}
	}
}

// render updates the game screen by drawing the map, GameState overlays,
// and other visual elements to the Display buffer.
func (e *Engine) render() {
	e.Display.BeginFrame()
	e.Display.Clear(core.Black) // Black background

	// Determine active theme based on global states
	activeTheme := e.BaseTheme

	if e.State == GameStatePaused {
		activeTheme = world.TileVariantPaused
	}

	e.renderMapLayer(activeTheme)
	systems.RenderEntities(e.EcsWorld, e.Display, e.Map, e.Camera)
	e.renderHUD()

	switch e.State {
	case GameStatePaused:
		e.renderPauseMenu()
	default:
	}

	e.Display.EndFrame()
}

func (e *Engine) renderPauseMenu() {
	e.drawTextCentered(14, "=== SYSTEM PAUSED ===", core.Red)
	e.drawTextCentered(16, "Press [ESC] to Resume", core.White)
	e.drawTextCentered(17, "Press [Q] to Quit", core.Gray)
}

func (e *Engine) renderMapLayer(theme world.TileVariant) {
	clear(e.PathLookup)

	// Collect paths from all PlayerControl entities to draw the red autopilot line
	targetMask := components.MaskPlayerControl
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			ctrl := e.EcsWorld.PlayerControls[i]
			if ctrl.Autopilot {
				for _, p := range ctrl.CurrentPath {
					e.PathLookup[p.Y*e.Map.Width+p.X] = true
				}
			}
		}
	}

	// Iterate over viewport bounds ONLY
	for y := e.Camera.Y; y < e.Camera.Y+e.Camera.Height; y++ {
		for x := e.Camera.X; x < e.Camera.X+e.Camera.Width; x++ {
			tile := e.Map.GetTile(x, y)
			if tile == nil {
				continue
			}

			// Translate world X,Y to screen coordinates
			screenX, screenY := e.Camera.WorldToScreen(x, y)

			// 1. Draw Autopilot Path
			isPathTile := e.PathLookup[y*e.Map.Width+x]
			if isPathTile && (tile.Visible || tile.Explored) {
				e.Display.DrawText(screenX, screenY, "*", core.Red)
				continue
			}

			// 2. Draw Map Base
			if tile.Type == world.TileTypeEmpty {
				continue
			}

			if tile.Visible {
				char, color := theme[tile.Type].Char, theme[tile.Type].Color

				if tile.Type == world.TileTypeFloor {
					char = "."
					// Subdue floor tile contrast so walls and entities pop out cleanly
					color = display.DarkenColor(color, 2)
				}

				if tile.Type == world.TileTypeWall {
					if char == "╬" || char == "#" || char == "█" || char == "▓" {
						switch tile.Bitmask {
						case 0: char = "O"
						case 1, 4, 5: char = "║"
						case 2, 8, 10: char = "═"
						case 3: char = "╚"
						case 6: char = "╔"
						case 12: char = "╗"
						case 9: char = "╝"
						case 7: char = "╠"
						case 14: char = "╦"
						case 13: char = "╣"
						case 11: char = "╩"
						case 15: char = "╬"
						}
					}
				}

				if tile.IsSunlit {
					sunColor := e.Clock.GetSunlightColor()
					color = core.LerpColor(color, sunColor, 0.75)
				}

				isTilePowered := systems.IsPowerActiveAt(e.EcsWorld, e.Map, x, y)
				isSunlitByDay := tile.IsSunlit && e.Clock.IsDaytime()
				if !isTilePowered && !isSunlitByDay {
					if tile.Distance > 3 { color = display.DarkenColor(color, 2) }
					if tile.Distance > 5 { color = display.DarkenColor(color, 2) }
				}

				e.Display.DrawText(screenX, screenY, char, color)
				continue
			}

			if tile.Explored {
				char, color := theme[tile.Type].Char, theme[tile.Type].Color
				if tile.Type == world.TileTypeFloor {
					char = "."
				}
				if tile.Type == world.TileTypeWall {
					if char == "╬" || char == "#" || char == "█" || char == "▓" {
						switch tile.Bitmask {
						case 0: char = "O"
						case 1, 4, 5: char = "║"
						case 2, 8, 10: char = "═"
						case 3: char = "╚"
						case 6: char = "╔"
						case 12: char = "╗"
						case 9: char = "╝"
						case 7: char = "╠"
						case 14: char = "╦"
						case 13: char = "╣"
						case 11: char = "╩"
						case 15: char = "╬"
						}
					}
				}
				dimColor := display.DarkenColor(color, 4)
				e.Display.DrawText(screenX, screenY, char, dimColor)
				continue
			}
		}
	}
}

func (e *Engine) renderHUD() {
	hudY := e.Camera.Height
	divider := strings.Repeat("═", e.Camera.Width)
	e.drawText(0, hudY, divider, core.Gray)

	statusText := "HEALTHY"
	autopilotEngaged := false
	var interactPrompt string

	targetMask := components.MaskPlayerControl | components.MaskPosition
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			control := e.EcsWorld.PlayerControls[i]
			position := e.EcsWorld.Positions[i]
			autopilotEngaged = control.Autopilot
			statusText = control.Status.Title()

			interactMask := components.MaskPosition | components.MaskInteractable
			for j := ecs.Entity(0); j < ecs.MaxEntities; j++ {
				if (e.EcsWorld.Masks[j] & interactMask) == interactMask {
					targetPos := e.EcsWorld.Positions[j]
					dx := targetPos.X - position.X
					dy := targetPos.Y - position.Y
					if (dx*dx + dy*dy) <= 2 {
						interact := e.EcsWorld.Interactables[j]
						interactPrompt = interact.Prompt
						break
					}
				}
			}
			break
		}
	}

	e.drawText(2, hudY+1, fmt.Sprintf(" STATUS: %s ", statusText), core.Cyan)
	if autopilotEngaged {
		e.drawText(22, hudY+1, "[ NAV-COM: AUTOPILOT ENGAGED ]", core.Red)
	} else {
		e.drawText(22, hudY+1, "[ NAV-COM: MANUAL OVERRIDE ]  ", core.Gray)
	}

	clockText := fmt.Sprintf(" TIME: %s ", e.Clock.FormatTime())
	clockX := e.Camera.Width - len(clockText) - 8
	if clockX < 55 {
		clockX = 55
	}
	e.drawText(clockX, hudY+1, clockText, core.Yellow)

	if interactPrompt != "" {
		if e.tickCount%30 < 15 {
			e.drawTextCentered(hudY-1, fmt.Sprintf("[ %s ]", interactPrompt), core.Green)
		}
	}

	for i, msg := range e.Messages {
		color := core.Green
		if i == 0 && len(e.Messages) == 3 { color = display.DarkenColor(core.Green, 3) }
		if i == 1 && len(e.Messages) == 3 { color = display.DarkenColor(core.Green, 1) }
		e.drawText(2, hudY+2+i, "> "+msg, color)
	}

	controls := " [W/A/S/D] Move    [P] Autopilot    [+/-] Zoom    [ESC] Pause    [Q] Abort"
	e.drawText(2, hudY+7, controls, core.Gray)
}

func (e *Engine) drawTextCentered(y int, text string, color core.Color) {
	centerX := e.Camera.Width / 2
	halfText := len(text) / 2
	x := centerX - halfText
	e.Display.DrawText(x, y, text, color)
}

func (e *Engine) drawText(x, y int, text string, color core.Color) {
	e.Display.DrawText(x, y, text, color)
}
