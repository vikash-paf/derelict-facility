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
	Camera        rl.Camera2D
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
		Camera: rl.Camera2D{
			Zoom:   1.0,
			Offset: rl.NewVector2(500, 300), // center of 1000x600 viewport
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

		// Update audio stream (music playback)
		e.Audio.Update()

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
		e.loadLevelByID(e.ActiveLevelID, 0, false) // false = arriving at start, use @ marker
		return
	}

	// Procedural fallback: generate a random map
	e.ActiveMission = nil
	e.ActiveLevelID = ""
	seed := time.Now().UnixNano()
	generator := world.NewFacilityGenerator(uint64(seed))
	loadedMap, playerX, playerY := generator.Generate(120, 40)
	e.activateMap(loadedMap, playerX, playerY, nil, "")
	e.State = GameStateRunning
}

// loadLevelByID loads a specific level from the active mission manifest, preserving
// the player's current security clearance across the transition.
// goingUp indicates whether the player arrived via an ascending (<) or descending (>) elevator,
// which determines which elevator in the destination they spawn next to.
func (e *Engine) loadLevelByID(levelID string, existingClearance uint32, goingUp bool) {
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
		loadedMap, defaultX, defaultY, err := loader.LoadBytes(data)
		if err != nil {
			e.Messages = append(e.Messages, fmt.Sprintf("ERROR: Cannot parse level %s", levelID))
			return
		}

		// Spawn player at the arrival elevator, not at the @ marker.
		// Going DOWN via > → arrive at the < elevator (IsUp=true) in destination.
		// Going UP via < → arrive at the > elevator (IsUp=false) in destination.
		spawnX, spawnY := arrivalSpawnPos(loadedMap, goingUp, defaultX, defaultY)

		e.ActiveLevelID = levelID
		e.activateMap(loadedMap, spawnX, spawnY, e.ActiveMission, levelID)
		if existingClearance != 0 {
			e.transferPlayerClearance(existingClearance)
		}
		e.State = GameStateRunning
		return
	}

	e.Messages = append(e.Messages, fmt.Sprintf("ERROR: Level %q not found in mission", levelID))
}

// arrivalSpawnPos finds the elevator in the destination map that the player steps out of.
// goingUp=true means player rode DOWN (>) so they arrive at the < elevator (IsUp=true).
// goingUp=false means player rode UP (<) so they arrive at the > elevator (IsUp=false).
// Falls back to the map's @ marker position if no matching elevator exists.
func arrivalSpawnPos(gameMap *world.Map, goingUp bool, fallbackX, fallbackY int) (int, int) {
	for _, s := range gameMap.Stairways {
		if s.IsUp == goingUp {
			// Spawn one tile away from the elevator so the player isn't on top of it
			return s.Pos.X + 1, s.Pos.Y
		}
	}
	return fallbackX, fallbackY
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
			e.Camera.Zoom += 0.25
			if e.Camera.Zoom > 3.0 {
				e.Camera.Zoom = 3.0
			}
			e.updateCamera()
		}
		if event.Key == rl.KeyMinus { // '-' key to Zoom Out
			e.Camera.Zoom -= 0.25
			if e.Camera.Zoom < 0.5 {
				e.Camera.Zoom = 0.5
			}
			e.updateCamera()
		}
		if event.Key == rl.KeyM {
			e.Audio.ToggleMute()
		}
		if event.Key == rl.KeyC {
			e.Display.ToggleShader()
		}
	}
}

func (e *Engine) recalculateViewport() {
	// No-op: Virtual resolution rendering handles scaling automatically.
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
	}, func(targetLevelID string, goingUp bool) {
		e.loadLevelByID(targetLevelID, clearanceBefore, goingUp)
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

func (e *Engine) getVisibleTileRange() (int, int, int, int) {
	if e.Map == nil {
		return 0, 0, 0, 0
	}
	cellWidth := float32(10)
	cellHeight := float32(20)
	viewW := float32(1000) / e.Camera.Zoom
	viewH := float32(600) / e.Camera.Zoom

	startX := int((e.Camera.Target.X - viewW/2) / cellWidth)
	endX := int((e.Camera.Target.X + viewW/2) / cellWidth) + 1
	startY := int((e.Camera.Target.Y - viewH/2) / cellHeight)
	endY := int((e.Camera.Target.Y + viewH/2) / cellHeight) + 1

	if startX < 0 { startX = 0 }
	if endX > e.Map.Width { endX = e.Map.Width }
	if startY < 0 { startY = 0 }
	if endY > e.Map.Height { endY = e.Map.Height }

	return startX, endX, startY, endY
}

func (e *Engine) updateAudioModulation() {
	targetMask := components.MaskPlayerControl | components.MaskPosition
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			pos := e.EcsWorld.Positions[i]
			isPowered := systems.IsPowerActiveAt(e.EcsWorld, e.Map, pos.X, pos.Y)
			if isPowered {
				e.Audio.SetMusicVolume(0.4)
				e.Audio.SetMusicPitch(1.0)
			} else {
				// Eerie unpowered room effect: lower volume, lower pitch
				e.Audio.SetMusicVolume(0.18)
				e.Audio.SetMusicPitch(0.82)
			}
			break
		}
	}
}

func (e *Engine) getPlayerPosition() (components.Position, bool) {
	targetMask := components.MaskPlayerControl | components.MaskPosition
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			return e.EcsWorld.Positions[i], true
		}
	}
	return components.Position{}, false
}

func (e *Engine) clampCameraToMap(cellWidth, cellHeight float32) {
	mapW := float32(e.Map.Width) * cellWidth
	mapH := float32(e.Map.Height) * cellHeight
	viewW := float32(1000) / e.Camera.Zoom
	viewH := float32(600) / e.Camera.Zoom

	if mapW > viewW {
		minX := viewW / 2
		maxX := mapW - viewW / 2
		if e.Camera.Target.X < minX {
			e.Camera.Target.X = minX
		} else if e.Camera.Target.X > maxX {
			e.Camera.Target.X = maxX
		}
	} else {
		e.Camera.Target.X = mapW / 2
	}

	if mapH > viewH {
		minY := viewH / 2
		maxY := mapH - viewH / 2
		if e.Camera.Target.Y < minY {
			e.Camera.Target.Y = minY
		} else if e.Camera.Target.Y > maxY {
			e.Camera.Target.Y = maxY
		}
	} else {
		e.Camera.Target.Y = mapH / 2
	}
}

func (e *Engine) updateCamera() {
	pos, found := e.getPlayerPosition()
	if !found {
		return
	}
	cellWidth := float32(10)
	cellHeight := float32(20)

	// Player position in pixels (center of player tile)
	playerPx := float32(pos.X)*cellWidth + cellWidth/2
	playerPy := float32(pos.Y)*cellHeight + cellHeight/2

	e.Camera.Target = rl.NewVector2(playerPx, playerPy)
	e.Camera.Offset = rl.NewVector2(500, 300) // Center of 1000x600 viewport

	// Clamp to map boundaries if map exists
	if e.Map != nil {
		e.clampCameraToMap(cellWidth, cellHeight)
	}
}

// render updates the game screen by drawing the map, GameState overlays,
// and other visual elements to the Display buffer.
func (e *Engine) render() {
	e.Display.BeginFrame()

	if e.State == GameStateMainMenu {
		e.Menu.Render(e.Display, 100, 38)
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
		startX, endX, startY, endY := e.getVisibleTileRange()
		bounds := core.ViewportBounds{
			StartX: startX,
			EndX:   endX,
			StartY: startY,
			EndY:   endY,
		}

		rl.BeginMode2D(e.Camera)
		e.renderMapLayer(activeTheme, startX, endX, startY, endY)
		systems.RenderEntities(e.EcsWorld, e.Display, e.Map, bounds)
		rl.EndMode2D()

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

func (e *Engine) populatePathLookup() {
	clear(e.PathLookup)
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
}

func (e *Engine) getFloorBackgroundColor(x, y int, tile *world.Tile) core.Color {
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
	return bgColor
}

func (e *Engine) getWallGlyphAndColor(x, y int, tile *world.Tile, theme world.TileVariant) (string, core.Color) {
	char, color := theme[tile.Type].Char, theme[tile.Type].Color
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
	return char, color
}

func (e *Engine) renderVisibleTile(x, y int, tile *world.Tile, theme world.TileVariant) {
	char, color := theme[tile.Type].Char, theme[tile.Type].Color

	if tile.Type == world.TileTypeFloor {
		bgColor := e.getFloorBackgroundColor(x, y, tile)
		e.Display.DrawRect(x, y, bgColor)
		return
	}

	if tile.Type == world.TileTypeWall {
		char, color = e.getWallGlyphAndColor(x, y, tile, theme)
	}

	isTilePowered := systems.IsPowerActiveAt(e.EcsWorld, e.Map, x, y)
	isSunlitByDay := tile.SunlightIntensity > 0.0 && e.Clock.IsDaytime()
	if !isTilePowered && !isSunlitByDay {
		if tile.Distance > 3 { color = display.DarkenColor(color, 2) }
		if tile.Distance > 5 { color = display.DarkenColor(color, 2) }
	}

	e.Display.DrawText(x, y, char, color)
}

func (e *Engine) renderExploredTile(x, y int, tile *world.Tile, theme world.TileVariant) {
	if tile.Type == world.TileTypeFloor {
		e.Display.DrawRect(x, y, core.Color{R: 8, G: 8, B: 12, A: 255})
		return
	}

	char, color := theme[tile.Type].Char, theme[tile.Type].Color
	if tile.Type == world.TileTypeWall {
		char, _ = e.getWallGlyphAndColor(x, y, tile, theme)
	}

	dimColor := display.DarkenColor(color, 4)
	e.Display.DrawText(x, y, char, dimColor)
}

func (e *Engine) renderSingleMapTile(x, y int, theme world.TileVariant) {
	tile := e.Map.GetTile(x, y)
	if tile == nil || tile.Type == world.TileTypeEmpty {
		return
	}

	// 1. Draw Autopilot Path
	if e.PathLookup[y*e.Map.Width+x] && (tile.Visible || tile.Explored) {
		e.Display.DrawText(x, y, "*", core.Red)
		return
	}

	// 2. Draw Visible Tiles
	if tile.Visible {
		e.renderVisibleTile(x, y, tile, theme)
		return
	}

	// 3. Draw Explored (Fog-of-War) Tiles
	if tile.Explored {
		e.renderExploredTile(x, y, tile, theme)
	}
}

func (e *Engine) renderMapLayer(theme world.TileVariant, startX, endX, startY, endY int) {
	e.populatePathLookup()

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			e.renderSingleMapTile(x, y, theme)
		}
	}
}

func (e *Engine) getPlayerControlAndPosition() (*components.PlayerControl, *components.Position, bool) {
	targetMask := components.MaskPlayerControl | components.MaskPosition
	for i := ecs.Entity(0); i < ecs.MaxEntities; i++ {
		if (e.EcsWorld.Masks[i] & targetMask) == targetMask {
			return &e.EcsWorld.PlayerControls[i], &e.EcsWorld.Positions[i], true
		}
	}
	return nil, nil, false
}

func (e *Engine) getNearbyInteractionPrompt(pX, pY int) string {
	interactMask := components.MaskPosition | components.MaskInteractable
	for j := ecs.Entity(0); j < ecs.MaxEntities; j++ {
		if (e.EcsWorld.Masks[j] & interactMask) == interactMask {
			targetPos := e.EcsWorld.Positions[j]
			dx := targetPos.X - pX
			dy := targetPos.Y - pY
			if (dx*dx + dy*dy) <= 2 {
				return e.EcsWorld.Interactables[j].Prompt
			}
		}
	}
	return ""
}

func (e *Engine) drawHUDStatusAndNav(hudY int, statusText string, autopilotEngaged bool) {
	e.drawText(2, hudY+1, fmt.Sprintf(" STATUS: %s ", statusText), core.Cyan)
	if autopilotEngaged {
		e.drawText(22, hudY+1, "[ NAV-COM: AUTOPILOT ENGAGED ]", core.Red)
	} else {
		e.drawText(22, hudY+1, "[ NAV-COM: MANUAL OVERRIDE ]  ", core.Gray)
	}
}

func (e *Engine) drawHUDMissionAndClock(hudY int) {
	missionText := " MISSION: PROCEDURAL "
	missionColor := core.Gray
	if e.ActiveMission != nil {
		missionText = fmt.Sprintf(" MISSION: %s ", e.ActiveMission.Title)
		missionColor = core.Yellow
	}
	e.drawText(2, hudY+2, missionText, missionColor)

	clockText := fmt.Sprintf(" TIME: %s ", e.Clock.FormatTime())
	clockX := 100 - len(clockText) - 8
	if clockX < 55 {
		clockX = 55
	}
	e.drawText(clockX, hudY+1, clockText, core.Yellow)
}

func (e *Engine) drawHUDInteractionPrompt(hudY int, interactPrompt string) {
	if interactPrompt != "" {
		if e.tickCount%30 < 15 {
			e.drawTextCentered(hudY-1, fmt.Sprintf("[ %s ]", interactPrompt), core.Green)
		}
	}
}

func (e *Engine) drawHUDMessages(hudY int) {
	for i, msg := range e.Messages {
		color := core.Green
		if i == 0 && len(e.Messages) == 3 { color = display.DarkenColor(core.Green, 3) }
		if i == 1 && len(e.Messages) == 3 { color = display.DarkenColor(core.Green, 1) }
		e.drawText(2, hudY+3+i, "> "+msg, color)
	}
}

func (e *Engine) drawHUDControls(hudY int) {
	muteLabel := "[M] Mute"
	controlsColor := core.Gray
	if e.Audio.IsMuted() {
		muteLabel = "[M] MUTED"
		controlsColor = core.Red
	}
	controls := fmt.Sprintf(" [W/A/S/D] Move    [P] Autopilot    [+/-] Zoom    %s    [C] CRT Filter    [ESC] Pause    [Q] Abort", muteLabel)
	e.drawText(2, hudY+7, controls, controlsColor)
}

func (e *Engine) renderHUD() {
	hudY := 30
	divider := strings.Repeat("═", 100)
	e.drawText(0, hudY, divider, core.Gray)

	ctrl, pos, found := e.getPlayerControlAndPosition()

	statusText := "HEALTHY"
	autopilotEngaged := false
	var interactPrompt string

	if found {
		statusText = ctrl.Status.Title()
		autopilotEngaged = ctrl.Autopilot
		interactPrompt = e.getNearbyInteractionPrompt(pos.X, pos.Y)
	}

	e.drawHUDStatusAndNav(hudY, statusText, autopilotEngaged)
	e.drawHUDMissionAndClock(hudY)
	e.drawHUDInteractionPrompt(hudY, interactPrompt)
	e.drawHUDMessages(hudY)
	e.drawHUDControls(hudY)
}

func (e *Engine) drawTextCentered(y int, text string, color core.Color) {
	centerX := 50
	halfText := len(text) / 2
	x := centerX - halfText
	e.Display.DrawText(x, y, text, color)
}

func (e *Engine) drawText(x, y int, text string, color core.Color) {
	e.Display.DrawText(x, y, text, color)
}
