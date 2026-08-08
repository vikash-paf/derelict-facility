package engine

import (
	"fmt"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/vikash-paf/derelict-facility/internal/audio"
	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/menu"
	"github.com/vikash-paf/derelict-facility/internal/mission"
	"github.com/vikash-paf/derelict-facility/internal/systems"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

const (
	fovRadius = 8 // cool stuff can be done here, like a dimming torch light
)

type GameState uint8

const (
	GameStateMainMenu GameState = iota
	GameStateRunning
	GameStatePaused
)

// Flip toggles between Paused and Running states
func (s GameState) Flip() GameState {
	if s == GameStatePaused {
		return GameStateRunning
	}
	if s == GameStateRunning {
		return GameStatePaused
	}
	return s
}

type Engine struct {
	Display       display.Display
	Audio         *audio.AudioManager
	Map           *world.Map
	EcsWorld      *ecs.World
	BaseTheme     world.TileVariant
	TickerRate    time.Duration
	tickCount     int
	Clock         *world.FacilityClock
	State         GameState
	Menu          *menu.MenuState
	Running       bool
	PathLookup    []bool // Pre-allocated array to avoid map allocations per frame
	Pathfinder    *world.Pathfinder
	Messages      []string
	Camera        *core.Camera
	ActiveMission *mission.MissionManifest // nil if using procedural map
	ActiveLevelID string                   // current level ID within the mission
}

func NewEngine(
	disp display.Display,
	gameMap *world.Map,
	ecsWorld *ecs.World,
	startingTheme world.TileVariant,
	viewWidth, viewHeight int,
) *Engine {
	audioMgr := audio.NewAudioManager()
	audioMgr.Init()

	e := &Engine{
		Display:    disp,
		Audio:      audioMgr,
		Map:        gameMap,
		EcsWorld:   ecsWorld,
		State:      GameStateMainMenu,
		Menu:       menu.NewMenuState(),
		Running:    true,
		TickerRate: time.Millisecond * 33, // ~30 fps
		BaseTheme:  startingTheme,
		Clock:      world.NewFacilityClock(),
		Camera: &core.Camera{
			X:      0,
			Y:      0,
			Width:  viewWidth,
			Height: viewHeight,
		},
	}

	if gameMap != nil {
		e.PathLookup = make([]bool, gameMap.Width*gameMap.Height)
		e.Pathfinder = world.NewPathfinder(gameMap.Width, gameMap.Height)
	}

	return e
}

// Run starts the deterministic game loop
func (e *Engine) Run() error {
	defer e.Audio.Close()

	for !e.Display.ShouldClose() && e.Running {
		if e.Display.IsResized() {
			e.recalculateViewport()
		}

		events := e.Display.PollInput()
		e.handleInputForGlobals(events)

		if e.State == GameStateMainMenu {
			e.handleMainMenuInput(events)
		} else if e.State == GameStateRunning {
			e.Update(events) // Calculate all game rules!
		}

		e.render() // Paint the results!
	}

	return nil
}

func (e *Engine) handleMainMenuInput(events []core.InputEvent) {
	for _, event := range events {
		switch event.Key {
		case rl.KeyUp:
			e.Menu.SelectPrevious()
			e.Audio.Play(audio.SoundFootstep)
		case rl.KeyDown:
			e.Menu.SelectNext()
			e.Audio.Play(audio.SoundFootstep)
		case rl.KeyEnter:
			e.Audio.Play(audio.SoundTerminalAccess)
			e.launchSelectedMission()
		}
	}
}

func (e *Engine) launchSelectedMission() {
	item := e.Menu.GetSelectedItem()

	if item.Manifest != nil {
		e.ActiveMission = item.Manifest
		e.ActiveLevelID = item.Manifest.StartLevel
		e.loadLevelByID(e.ActiveLevelID, 0)
		return
	}

	// Procedural fallback: generate a random map
	e.ActiveMission = nil
	e.ActiveLevelID = ""
	seed := time.Now().UnixNano()
	generator := world.NewFacilityGenerator(uint64(seed))
	loadedMap, playerX, playerY := generator.Generate(120, 40)
	e.activateMap(loadedMap, playerX, playerY, nil, "")
}

// loadLevelByID loads a specific level from the active mission manifest, preserving
// the player's current security clearance across the transition.
func (e *Engine) loadLevelByID(levelID string, existingClearance uint32) {
	if e.ActiveMission == nil {
		return
	}

	for _, lvl := range e.ActiveMission.Levels {
		if lvl.ID != levelID {
			continue
		}

		data, err := e.ActiveMission.LoadLevelMapData(lvl.File)
		if err != nil {
			e.Messages = append(e.Messages, fmt.Sprintf("ERROR: Cannot load level %s", levelID))
			return
		}

		loader := world.NewJSONMapLoader()
		loadedMap, playerX, playerY, err := loader.LoadBytes(data)
		if err != nil {
			e.Messages = append(e.Messages, fmt.Sprintf("ERROR: Cannot parse level %s", levelID))
			return
		}

		e.ActiveLevelID = levelID
		e.activateMap(loadedMap, playerX, playerY, e.ActiveMission, levelID)
		if existingClearance != 0 {
			e.transferPlayerClearance(existingClearance)
		}
		e.State = GameStateRunning
		return
	}

	e.Messages = append(e.Messages, fmt.Sprintf("ERROR: Level %q not found in mission", levelID))
}

// activateMap swaps out the engine's active map, rebuilds pathfinding, and constructs the ECS world.
func (e *Engine) activateMap(gameMap *world.Map, playerX, playerY int, manifest *mission.MissionManifest, levelID string) {
	e.Map = gameMap
	e.PathLookup = make([]bool, gameMap.Width*gameMap.Height)
	e.Pathfinder = world.NewPathfinder(gameMap.Width, gameMap.Height)
	e.EcsWorld = buildMissionWorld(gameMap, playerX, playerY, manifest, levelID)
	e.Messages = nil
}

// transferPlayerClearance copies clearance bits from a previous level into the current ECS world player.
func (e *Engine) transferPlayerClearance(clearance uint32) {
	targetMask := components.MaskPlayerControl
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			e.EcsWorld.PlayerControls[i].SecurityClearance = clearance
			return
		}
	}
}

// playerClearance returns the current player's security clearance bitmask.
func (e *Engine) playerClearance() uint32 {
	targetMask := components.MaskPlayerControl
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			return e.EcsWorld.PlayerControls[i].SecurityClearance
		}
	}
	return 0
}

// buildMissionWorld constructs the ECS world for a given map.
// manifest and levelID are optional — if provided, stairway entities are given
// their correct TargetLevelID based on position within the level sequence.
func buildMissionWorld(gameMap *world.Map, playerX, playerY int, manifest *mission.MissionManifest, levelID string) *ecs.World {
	ecsWorld := ecs.NewWorld()

	playerEnt := ecsWorld.CreateEntity()
	ecsWorld.AddPosition(playerEnt, components.Position{X: playerX, Y: playerY})
	ecsWorld.AddGlyph(playerEnt, components.Glyph{Char: "@", Color: core.BrightWhite})
	ecsWorld.AddPlayerControl(playerEnt, components.PlayerControl{
		Autopilot: false,
		Status:    components.PlayerStatusHealthy,
	})

	spawnGenerators(ecsWorld, gameMap)
	spawnTerminals(ecsWorld, gameMap)
	spawnDoors(ecsWorld, gameMap, playerX, playerY)
	spawnStairways(ecsWorld, gameMap, manifest, levelID)

	return ecsWorld
}

func spawnGenerators(w *ecs.World, gameMap *world.Map) {
	for _, genInfo := range gameMap.PowerGenerators {
		genEnt := w.CreateEntity()
		w.AddPosition(genEnt, components.Position{X: genInfo.Pos.X, Y: genInfo.Pos.Y})
		w.AddGlyph(genEnt, components.Glyph{Char: "X", Color: core.Red})
		w.AddSolid(genEnt)
		w.AddInteractable(genEnt, components.Interactable{Prompt: "Press [E] to Toggle Generator"})
		w.AddPowerGenerator(genEnt, components.PowerGenerator{
			IsActive: false,
			IsGlobal: genInfo.IsGlobal,
		})
	}
}

func spawnTerminals(w *ecs.World, gameMap *world.Map) {
	for _, termPos := range gameMap.Terminals {
		termEnt := w.CreateEntity()
		w.AddPosition(termEnt, components.Position{X: termPos.X, Y: termPos.Y})
		w.AddGlyph(termEnt, components.Glyph{Char: "T", Color: core.Cyan})
		w.AddSolid(termEnt)
		w.AddInteractable(termEnt, components.Interactable{Prompt: "Press [E] to Access Terminal"})
		w.AddTerminal(termEnt, components.Terminal{HasSaved: false})
		w.AddNarrative(termEnt, components.Narrative{
			Text: "LOG: Sector 4 containment breach. All non-essential personnel evacuate immediately.",
		})
	}
}

func spawnDoors(w *ecs.World, gameMap *world.Map, playerX, playerY int) {
	for _, doorPos := range gameMap.Doors {
		if doorPos.X == playerX && doorPos.Y == playerY {
			continue
		}

		doorEnt := w.CreateEntity()
		w.AddPosition(doorEnt, components.Position{X: doorPos.X, Y: doorPos.Y})
		w.AddGlyph(doorEnt, components.Glyph{Char: "+", Color: core.White})
		w.AddSolid(doorEnt)
		w.AddInteractable(doorEnt, components.Interactable{Prompt: "Press [E] to Open Door"})
		w.AddDoor(doorEnt, components.Door{IsOpen: false})
	}
}

// spawnStairways creates stairway entities and resolves target level IDs from the manifest level sequence.
func spawnStairways(w *ecs.World, gameMap *world.Map, manifest *mission.MissionManifest, levelID string) {
	if len(gameMap.Stairways) == 0 {
		return
	}

	// Build prev/next level IDs from the manifest sequence
	prevLevelID, nextLevelID := resolveLevelNeighbors(manifest, levelID)

	for _, stairInfo := range gameMap.Stairways {
		targetID := nextLevelID
		if stairInfo.IsUp {
			targetID = prevLevelID
		}

		dirGlyph, dirColor := ">", core.Yellow
		prompt := "Press [E] to Descend"
		if stairInfo.IsUp {
			dirGlyph, dirColor = "<", core.Cyan
			prompt = "Press [E] to Ascend"
		}

		stairEnt := w.CreateEntity()
		w.AddPosition(stairEnt, components.Position{X: stairInfo.Pos.X, Y: stairInfo.Pos.Y})
		w.AddGlyph(stairEnt, components.Glyph{Char: dirGlyph, Color: dirColor})
		w.AddInteractable(stairEnt, components.Interactable{Prompt: prompt})
		w.AddStairway(stairEnt, components.Stairway{
			TargetLevelID: targetID,
			IsUp:          stairInfo.IsUp,
		})
	}
}

// resolveLevelNeighbors returns the previous and next level IDs for the given level in a manifest.
func resolveLevelNeighbors(manifest *mission.MissionManifest, levelID string) (prev, next string) {
	if manifest == nil {
		return "", ""
	}
	for i, lvl := range manifest.Levels {
		if lvl.ID != levelID {
			continue
		}
		if i > 0 {
			prev = manifest.Levels[i-1].ID
		}
		if i < len(manifest.Levels)-1 {
			next = manifest.Levels[i+1].ID
		}
		return
	}
	return
}

func (e *Engine) handleInputForGlobals(events []core.InputEvent) {
	for _, event := range events {
		if event.Quit || event.Key == rl.KeyQ {
			e.Running = false
			return
		}
		if event.Key == rl.KeyEscape && e.State != GameStateMainMenu {
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
	// Capture clearance before processing so we can transfer it on level change
	clearanceBefore := e.playerClearance()

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
	}, func(soundID string) {
		e.Audio.Play(audio.SoundID(soundID))
	}, func(targetLevelID string) {
		e.loadLevelByID(targetLevelID, clearanceBefore)
	})

	// Center camera on player
	e.updateCamera()

	// Run AI movement every 2nd frame (approx 15 times a second)
	if e.tickCount%6 == 0 {
		systems.ProcessAutopilot(e.EcsWorld, e.Map, e.Pathfinder, func(soundID string) {
			e.Audio.Play(audio.SoundID(soundID))
		})
	}

	// Calculate sunlight spillover through open doors/corridors
	e.Map.PropagateSunlight(3, func(x, y int) bool {
		return systems.IsSolidAt(e.EcsWorld, x, y)
	})

	// Recalculate FOV based on current player position
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		targetMask := components.MaskPlayerControl | components.MaskPosition
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			pos := e.EcsWorld.Positions[i]
			e.Map.ComputeFOV(pos.X, pos.Y, fovRadius, func(x, y int) bool {
				if !e.Map.IsWalkable(x, y) {
					return true
				}
				return systems.IsSolidAt(e.EcsWorld, x, y)
			}, func(x, y int) bool {
				if systems.IsPowerActiveAt(e.EcsWorld, e.Map, x, y) {
					return true
				}
				tile := e.Map.GetTile(x, y)
				return tile != nil && tile.SunlightIntensity > 0.0 && e.Clock.IsDaytime()
			})
			break
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
			newX := pos.X - e.Camera.Width/2
			newY := pos.Y - e.Camera.Height/2

			if newX < 0 {
				newX = 0
			}
			if newY < 0 {
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

	if e.State == GameStateMainMenu {
		e.Menu.Render(e.Display, e.Camera.Width, e.Camera.Height+8)
		e.Display.EndFrame()
		return
	}

	e.Display.Clear(core.Black) // Black background

	// Determine active theme based on global states
	activeTheme := e.BaseTheme

	if e.State == GameStatePaused {
		activeTheme = world.TileVariantPaused
	}

	if e.Map != nil && e.EcsWorld != nil {
		e.renderMapLayer(activeTheme)
		systems.RenderEntities(e.EcsWorld, e.Display, e.Map, e.Camera)
		e.renderHUD()
	}

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

				// Draw floor tile background fill using ambient daylight color
				if tile.Type == world.TileTypeFloor {
					bgColor := core.Color{R: 20, G: 20, B: 25, A: 255} // base floor dark fill
					sunColor := e.Clock.GetSunlightColor()
					if tile.SunlightIntensity > 0.0 && e.Clock.IsDaytime() {
						blendWeight := 0.40 * tile.SunlightIntensity
						bgColor = core.LerpColor(bgColor, sunColor, blendWeight)
					}

					isTilePowered := systems.IsPowerActiveAt(e.EcsWorld, e.Map, x, y)
					isSunlitByDay := tile.SunlightIntensity > 0.0 && e.Clock.IsDaytime()
					if !isTilePowered && !isSunlitByDay {
						if tile.Distance > 3 { bgColor = display.DarkenColor(bgColor, 2) }
						if tile.Distance > 5 { bgColor = display.DarkenColor(bgColor, 2) }
					}

					e.Display.DrawRect(screenX, screenY, bgColor)
					continue
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

					sunColor := e.Clock.GetSunlightColor()
					if tile.SunlightIntensity > 0.0 && e.Clock.IsDaytime() {
						blendWeight := 0.45 * tile.SunlightIntensity
						color = core.LerpColor(color, sunColor, blendWeight)
					}
				}

				isTilePowered := systems.IsPowerActiveAt(e.EcsWorld, e.Map, x, y)
				isSunlitByDay := tile.SunlightIntensity > 0.0 && e.Clock.IsDaytime()
				if !isTilePowered && !isSunlitByDay {
					if tile.Distance > 3 { color = display.DarkenColor(color, 2) }
					if tile.Distance > 5 { color = display.DarkenColor(color, 2) }
				}

				e.Display.DrawText(screenX, screenY, char, color)
				continue
			}

			if tile.Explored {
				if tile.Type == world.TileTypeFloor {
					e.Display.DrawRect(screenX, screenY, core.Color{R: 8, G: 8, B: 12, A: 255})
					continue
				}
				char, color := theme[tile.Type].Char, theme[tile.Type].Color
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
