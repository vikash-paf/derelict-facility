package engine

import (
	"fmt"
	"math"
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
	fovRadius  = 8 // cool stuff can be done here, like a dimming torch light
	cellWidth  = 20
	cellHeight = 40
	hudHeight  = 8
	viewportW  = 2000
	viewportH  = 1200
	virtualW   = 100
	virtualH   = 38
	hudY       = 30
)

type GameState uint8

const (
	GameStateMainMenu GameState = iota
	GameStateRunning
	GameStatePaused
	GameStateGameOver
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


type LevelState struct {
	Map      *world.Map
	EcsWorld *ecs.World
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
	LevelCache    map[string]*LevelState
	heartbeatTimer float64
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
		LevelCache: make(map[string]*LevelState),
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
		} else if e.State == GameStateGameOver {
			e.handleGameOverInput(events)
		} else if e.State == GameStateRunning {
			e.Update(events) // Calculate all game rules!
		}

		e.render() // Paint the results!
	}

	return nil
}

func (e *Engine) handleGameOverInput(events []core.InputEvent) {
	for _, event := range events {
		if event.Key == rl.KeyR {
			// Restore savegame
			saveData, err := systems.LoadState()
			if err == nil {
				e.EcsWorld = saveData.World
				e.Map = saveData.Map
				e.ActiveLevelID = saveData.ActiveLevelID
				e.ActiveMission = saveData.ActiveMission
				e.Clock.TotalTicks = saveData.TotalTicks
				e.Clock.Day = saveData.Day
				e.Clock.Season = world.Season(saveData.Season)
				
				e.PathLookup = make([]bool, e.Map.Width*e.Map.Height)
				e.Pathfinder = world.NewPathfinder(e.Map.Width, e.Map.Height)
				e.LevelCache = make(map[string]*LevelState) // reset level caches
				
				e.Messages = []string{"Checkpoint restored successfully."}
				e.State = GameStateRunning
				fmt.Println("Checkpoint restored successfully from savegame.sav")
			} else {
				errMsg := fmt.Sprintf("Load checkpoint failed: %v", err)
				e.Messages = []string{errMsg}
				fmt.Println(errMsg)
			}
		}
		if event.Key == rl.KeyM {
			e.State = GameStateMainMenu
		}
	}
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
		e.loadLevelByID(e.ActiveLevelID, 0, false, true) // true = starting fresh, spawn at @
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

func (e *Engine) loadLevelByID(levelID string, existingClearance uint32, goingUp bool, isStart bool) {

	if e.ActiveMission == nil {
		return
	}

	// 1. Cache the current level state before transitioning
	if e.ActiveLevelID != "" && e.Map != nil && e.EcsWorld != nil {
		// Clean up the player's control path to avoid moving on arrival
		for i := range ecs.Entity(ecs.MaxEntities) {
			if e.EcsWorld.IsPlayer(i) {
				e.EcsWorld.PlayerControls[i].CurrentPath = nil
				e.EcsWorld.PlayerControls[i].Autopilot = false
				break
			}
		}
		e.LevelCache[e.ActiveLevelID] = &LevelState{
			Map:      e.Map,
			EcsWorld: e.EcsWorld,
		}
	}

	// 2. Try loading the level state from cache
	if cached, exists := e.LevelCache[levelID]; exists && !isStart {
		// Level is already cached. Restore it!
		e.ActiveLevelID = levelID
		e.Map = cached.Map
		e.EcsWorld = cached.EcsWorld
		e.PathLookup = make([]bool, e.Map.Width*e.Map.Height)
		e.Pathfinder = world.NewPathfinder(e.Map.Width, e.Map.Height)
		e.Messages = nil

		// Find player start or elevator spawn coordinate
		var spawnX, spawnY int
		// Resolve arrival coordinate on cached map
		for _, lvl := range e.ActiveMission.Levels {
			if lvl.ID == levelID {
				data, err := e.ActiveMission.LoadLevelMapData(lvl.File)
				if err == nil {
					loader := world.NewJSONMapLoader()
					if _, defX, defY, err := loader.LoadBytes(data); err == nil {
						spawnX, spawnY = arrivalSpawnPos(e.Map, goingUp, defX, defY)
					}
				}
				break
			}
		}

		// Update cached player entity coordinates to the new elevator entry point
		for i := range ecs.Entity(ecs.MaxEntities) {
			if e.EcsWorld.IsPlayer(i) {
				e.EcsWorld.Positions[i] = components.Position{X: spawnX, Y: spawnY}
				break
			}
		}

		if existingClearance != 0 {
			e.transferPlayerClearance(existingClearance)
		}
		e.State = GameStateRunning
		return
	}

	// 3. Fallback: Parse from map JSON file (First visit)
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

		// Spawn player at the map's start position (@ marker) if starting fresh,
		// otherwise spawn next to the correct elevator arrival stairway.
		var spawnX, spawnY int
		if isStart {
			spawnX, spawnY = defaultX, defaultY
		} else {
			spawnX, spawnY = arrivalSpawnPos(loadedMap, goingUp, defaultX, defaultY)
		}

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
// goingUp=true means player rode UP (<) so they arrive at the > elevator (IsUp=false).
// goingUp=false means player rode DOWN (>) so they arrive at the < elevator (IsUp=true).
// Falls back to the map's @ marker position if no matching elevator exists.
func arrivalSpawnPos(gameMap *world.Map, goingUp bool, fallbackX, fallbackY int) (int, int) {
	for _, s := range gameMap.Stairways {
		if s.IsUp == !goingUp {
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
	for i := range ecs.Entity(ecs.MaxEntities) {

		if e.EcsWorld.IsPlayer(i) {
			e.EcsWorld.PlayerControls[i].SecurityClearance = clearance
			return
		}
	}
}

// playerClearance returns the current player's security clearance bitmask.
func (e *Engine) playerClearance() uint32 {
	for i := range ecs.Entity(ecs.MaxEntities) {
		if e.EcsWorld.IsPlayer(i) {
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
		Survival: components.PlayerSurvival{
			Oxygen:    100.0,
			Toxicity:  0.0,
			MaxOxygen: 100.0,
			Health:    100.0,
			MaxHealth: 100.0,
		},
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
		w.AddInteractable(genEnt, components.Interactable{Prompt: "Press [E] to Turn On Generator"})
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
		e.updateHeartbeat(0.033) // ~30 fps delta time is 33ms
	}
}

func (e *Engine) updateHeartbeat(dt float64) {
	ctrl, _, found := e.getPlayerControlAndPosition()
	if !found {
		return
	}

	healthPct := ctrl.Survival.Health / 100.0
	toxPct := ctrl.Survival.Toxicity / 100.0

	// Trigger audio warnings if health drops below 35% or toxicity goes above 40%
	isCritical := healthPct < 0.35 || toxPct > 0.40
	if !isCritical {
		e.heartbeatTimer = 0
		return
	}

	severity := 0.0
	if healthPct < 0.35 {
		severity = math.Max(severity, (0.35-healthPct)/0.35)
	}
	if toxPct > 0.40 {
		severity = math.Max(severity, (toxPct-0.40)/0.60)
	}

	// Dynamic tempo mapping: 1.5 seconds down to 0.4 seconds pacing speed
	minInterval := 0.4
	maxInterval := 1.5
	interval := maxInterval - (severity * (maxInterval - minInterval))

	e.heartbeatTimer += dt
	if e.heartbeatTimer >= interval {
		e.Audio.Play(audio.SoundHeartbeat)
		e.heartbeatTimer = 0.0
	}
}


func (e *Engine) processSimulation(events []core.InputEvent) {
	// Capture clearance before processing so we can transfer it on level change
	clearanceBefore := e.playerClearance()

	// Let the systems tick using the events we polled at the start of the frame!
	systems.ProcessPlayerInput(e.EcsWorld, events, e.Map, e.ActiveMission, e.ActiveLevelID, func(msg string) {
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
		e.loadLevelByID(targetLevelID, clearanceBefore, goingUp, false)
	}, func() {
		// Exported save callback triggers SaveState
		systems.SaveState(e.EcsWorld, e.Map, e.ActiveMission, e.ActiveLevelID, e.Clock.TotalTicks, e.Clock.Day, uint8(e.Clock.Season))
	})


	// Run Life Support tick (deplete O2/gain Toxicity) and Hydroponics growth tick
	systems.ProcessLifeSupport(e.EcsWorld, e.Map, e.Clock, func(msg string) {
		if len(e.Messages) > 0 && e.Messages[len(e.Messages)-1] == msg {
			return
		}
		e.Messages = append(e.Messages, msg)
		if len(e.Messages) > 3 {
			e.Messages = e.Messages[1:]
		}
	})

	systems.ProcessHydroponics(e.EcsWorld, e.Map, e.Clock)

	// Check if player health reaches 0 (Game Over)
	for i := range ecs.Entity(ecs.MaxEntities) {
		if e.EcsWorld.IsPlayer(i) {
			if e.EcsWorld.PlayerControls[i].Survival.Health <= 0.0 {
				e.Messages = append(e.Messages, "CRITICAL ERROR: BIOSPHERE ENGINEER ELIMINATED.")
				e.State = GameStateGameOver
			}
			break
		}
	}


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
	for i := range ecs.Entity(ecs.MaxEntities) {
		if e.EcsWorld.IsPlayer(i) && e.EcsWorld.HasPosition(i) {
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
	cellWidthVal := float32(cellWidth)
	cellHeightVal := float32(cellHeight)
	viewW := float32(viewportW) / e.Camera.Zoom
	viewH := float32(viewportH) / e.Camera.Zoom

	startX := int((e.Camera.Target.X - viewW/2) / cellWidthVal)
	endX := int((e.Camera.Target.X+viewW/2)/cellWidthVal) + 1
	startY := int((e.Camera.Target.Y - viewH/2) / cellHeightVal)
	endY := int((e.Camera.Target.Y+viewH/2)/cellHeightVal) + 1

	if startX < 0 {
		startX = 0
	}
	if endX > e.Map.Width {
		endX = e.Map.Width
	}
	if startY < 0 {
		startY = 0
	}
	if endY > e.Map.Height {
		endY = e.Map.Height
	}

	return startX, endX, startY, endY
}

func (e *Engine) updateAudioModulation() {
	if e.EcsWorld == nil {
		return
	}
	for i := range ecs.Entity(ecs.MaxEntities) {
		if e.EcsWorld.IsPlayer(i) && e.EcsWorld.HasPosition(i) {
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
	if e.EcsWorld == nil {
		return components.Position{}, false
	}
	for i := range ecs.Entity(ecs.MaxEntities) {
		if e.EcsWorld.IsPlayer(i) && e.EcsWorld.HasPosition(i) {
			return e.EcsWorld.Positions[i], true
		}
	}
	return components.Position{}, false
}

func (e *Engine) clampCameraToMap(cellWidthVal, cellHeightVal float32) {
	mapW := float32(e.Map.Width) * cellWidthVal
	mapH := float32(e.Map.Height) * cellHeightVal
	viewW := float32(viewportW) / e.Camera.Zoom
	viewH := float32(viewportH) / e.Camera.Zoom

	if mapW > viewW {
		minX := viewW / 2
		maxX := mapW - viewW/2
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
		maxY := mapH - viewH/2
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
	cellWidthVal := float32(cellWidth)
	cellHeightVal := float32(cellHeight)

	// Player position in pixels (center of player tile)
	playerPx := float32(pos.X)*cellWidthVal + cellWidthVal/2
	playerPy := float32(pos.Y)*cellHeightVal + cellHeightVal/2

	e.Camera.Target = rl.NewVector2(playerPx, playerPy)
	e.Camera.Offset = rl.NewVector2(float32(viewportW)/2, float32(viewportH)/2)

	// Clamp to map boundaries if map exists
	if e.Map != nil {
		e.clampCameraToMap(cellWidthVal, cellHeightVal)
	}

	// Apply disorientation wobble/sway if player has SICK status (from toxicity)
	ctrl, _, controlFound := e.getPlayerControlAndPosition()
	if controlFound && ctrl.Status == components.PlayerStatusSick {
		t := float32(rl.GetTime())
		// Wobble camera target position with wave combinations
		e.Camera.Target.X += float32(math.Sin(float64(t*2.0))) * 12.0
		e.Camera.Target.Y += float32(math.Cos(float64(t*1.5))) * 12.0
		// Apply slight camera rotational roll (tilt)
		e.Camera.Rotation = float32(math.Sin(float64(t*1.0))) * 2.5
	} else {
		e.Camera.Rotation = 0.0
	}
}

// render updates the game screen by drawing the map, GameState overlays,
// and other visual elements to the Display buffer.
func (e *Engine) render() {
	e.Display.BeginFrame()

	if e.State == GameStateMainMenu {
		e.Menu.Render(e.Display, virtualW, virtualH)
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
	case GameStateGameOver:
		e.renderGameOver()
	default:
	}

	e.Display.EndFrame()
}

func (e *Engine) renderPauseMenu() {
	e.drawTextCentered(14, "=== SYSTEM PAUSED ===", core.Red)
	e.drawTextCentered(16, "Press [ESC] to Resume", core.White)
	e.drawTextCentered(17, "Press [Q] to Quit", core.Gray)
}

func (e *Engine) renderGameOver() {
	e.drawTextCentered(13, "=== CRITICAL FAILURE ===", core.Red)
	e.drawTextCentered(15, "BIOSPHERE MAINTENANCE ENGINEER ELIMINATED", core.BrightWhite)
	e.drawTextCentered(17, "Press [R] to Restore Checkpoint", core.Yellow)
	e.drawTextCentered(18, "Press [M] to Abort to Main Menu", core.Gray)
}


func (e *Engine) populatePathLookup() {
	clear(e.PathLookup)
	for i := range ecs.Entity(ecs.MaxEntities) {
		if e.EcsWorld.IsPlayer(i) {
			ctrl := e.EcsWorld.PlayerControls[i]
			if ctrl.Autopilot {
				for _, p := range ctrl.CurrentPath {
					e.PathLookup[p.Y*e.Map.Width+p.X] = true
				}
			}
		}
	}
}

func (e *Engine) getFlickerMultiplier(x, y int) float32 {
	t := float64(e.tickCount)
	// Compute offset wave patterns based on coordinate and time to differentiate rooms
	wave := math.Sin(t*0.25+float64(x)*0.3) * math.Cos(t*0.18+float64(y)*0.4)
	wave = (wave + 1.0) * 0.5

	// Sudden emergency power sparks or dim glow
	if wave > 0.88 {
		return 0.15 // sudden darkness drop
	}
	if wave > 0.75 {
		return 0.50 // dim light pulse
	}
	return 1.0 // full ambient light
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
		// Apply backup flickering FX to emergency lights
		flicker := e.getFlickerMultiplier(x, y)
		bgColor = core.Color{
			R: uint8(float32(bgColor.R) * flicker),
			G: uint8(float32(bgColor.G) * flicker),
			B: uint8(float32(bgColor.B) * flicker),
			A: bgColor.A,
		}

		if tile.Distance > 3 {
			bgColor = display.DarkenColor(bgColor, 2)
		}
		if tile.Distance > 5 {
			bgColor = display.DarkenColor(bgColor, 2)
		}
	}
	return bgColor
}


func (e *Engine) getWallGlyphAndColor(x, y int, tile *world.Tile, theme world.TileVariant) (string, core.Color) {
	char, color := theme[tile.Type].Char, theme[tile.Type].Color
	if char == "╬" || char == "#" || char == "█" || char == "▓" {
		switch tile.Bitmask {
		case world.WallBitmaskIsolated:
			char = "O"
		case world.WallBitmaskNorthOnly, world.WallBitmaskSouthOnly, world.WallBitmaskNorthSouth:
			char = "║"
		case world.WallBitmaskEastOnly, world.WallBitmaskWestOnly, world.WallBitmaskEastWest:
			char = "═"
		case world.WallBitmaskNorthEast:
			char = "╚"
		case world.WallBitmaskEastSouth:
			char = "╔"
		case world.WallBitmaskSouthWest:
			char = "╗"
		case world.WallBitmaskWestNorth:
			char = "╝"
		case world.WallBitmaskNorthEastSouth:
			char = "╠"
		case world.WallBitmaskEastSouthWest:
			char = "╦"
		case world.WallBitmaskSouthWestNorth:
			char = "╣"
		case world.WallBitmaskWestNorthEast:
			char = "╩"
		case world.WallBitmaskIntersection:
			char = "╬"
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
	if tile.Type == world.TileTypeFloor {
		bgColor := e.getFloorBackgroundColor(x, y, tile)
		e.Display.DrawRect(x, y, bgColor)

		// Draw plant overlay if present
		if tile.PlantStage != world.PlantStageNone {
			char, color := "h", core.Color{R: 50, G: 120, B: 50, A: 255}
			if tile.PlantStage == world.PlantStageSprout {
				char, color = "y", core.Color{R: 120, G: 200, B: 120, A: 255}
			} else if tile.PlantStage == world.PlantStageMature {
				char, color = "H", core.Color{R: 50, G: 240, B: 50, A: 255}
				// Use different colors to distinguish yields (O2 vs Medpack)
				if tile.YieldItemType == "MEDPACK" {
					color = core.Color{R: 240, G: 50, B: 240, A: 255} // purple-magenta for medical
				}
			}
			
			// Highlight mature plants with bright yellow interaction hint if adjacent to player
			_, pos, found := e.getPlayerControlAndPosition()
			if found && tile.PlantStage == world.PlantStageMature {
				dx := pos.X - x
				dy := pos.Y - y
				if dx*dx + dy*dy <= 2 {
					if e.tickCount%30 < 15 {
						color = core.Yellow
					}
				}
			}

			e.Display.DrawText(x, y, char, color)
		}
		return
	}

	char, color := theme[tile.Type].Char, theme[tile.Type].Color
	if tile.Type == world.TileTypeWall {
		char, color = e.getWallGlyphAndColor(x, y, tile, theme)
	}

	isTilePowered := systems.IsPowerActiveAt(e.EcsWorld, e.Map, x, y)
	isSunlitByDay := tile.SunlightIntensity > 0.0 && e.Clock.IsDaytime()
	if !isTilePowered && !isSunlitByDay {
		if tile.Distance > 3 {
			color = display.DarkenColor(color, 2)
		}
		if tile.Distance > 5 {
			color = display.DarkenColor(color, 2)
		}
	}

	e.Display.DrawText(x, y, char, color)
}

func (e *Engine) renderExploredTile(x, y int, tile *world.Tile, theme world.TileVariant) {
	if tile.Type == world.TileTypeFloor {
		e.Display.DrawRect(x, y, core.Color{R: 8, G: 8, B: 12, A: 255})
		
		// Render dim plants in fog of war
		if tile.PlantStage != world.PlantStageNone {
			char := "h"
			if tile.PlantStage == world.PlantStageSprout {
				char = "y"
			} else if tile.PlantStage == world.PlantStageMature {
				char = "H"
			}
			dimColor := core.Color{R: 20, G: 50, B: 20, A: 255}
			e.Display.DrawText(x, y, char, dimColor)
		}
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
	for i := range ecs.Entity(ecs.MaxEntities) {
		if e.EcsWorld.IsPlayer(i) && e.EcsWorld.HasPosition(i) {
			return &e.EcsWorld.PlayerControls[i], &e.EcsWorld.Positions[i], true
		}
	}
	return nil, nil, false
}

func (e *Engine) getNearbyInteractionPrompt(pX, pY int) string {
	// First check adjacent map-tile plants
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			tx := pX + dx
			ty := pY + dy
			tile := e.Map.GetTile(tx, ty)
			if tile != nil && tile.PlantStage == world.PlantStageMature {
				return "Press [E] to Harvest Crop"
			}
		}
	}

	for j := range ecs.Entity(ecs.MaxEntities) {
		if e.EcsWorld.HasPosition(j) && e.EcsWorld.IsInteractable(j) {
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


func (e *Engine) drawHUDStatusAndNav(hudY int, statusText string, autopilotEngaged bool, survival components.PlayerSurvival) {
	e.drawText(2, hudY+1, fmt.Sprintf("STATUS: %-11s", statusText), core.Cyan)

	// Draw Health, O2, and Toxicity Gauges shifted right to avoid overlaps
	hpFill := int(survival.Health / 10.0)
	o2Fill := int(survival.Oxygen / 10.0)
	toxFill := int(survival.Toxicity / 10.0)

	hpBar := strings.Repeat("█", hpFill) + strings.Repeat("░", 10-hpFill)
	o2Bar := strings.Repeat("█", o2Fill) + strings.Repeat("░", 10-o2Fill)
	toxBar := strings.Repeat("█", toxFill) + strings.Repeat("░", 10-toxFill)

	e.drawText(24, hudY+1, fmt.Sprintf("HP [%s] %3.0f%%", hpBar, survival.Health), core.Green)
	
	o2Color := core.Green
	if survival.Oxygen < 30.0 {
		o2Color = core.Red
	} else if survival.Oxygen < 60.0 {
		o2Color = core.Yellow
	}
	e.drawText(47, hudY+1, fmt.Sprintf("O2 [%s] %3.0f%%", o2Bar, survival.Oxygen), o2Color)

	toxColor := core.Gray
	if survival.Toxicity > 50.0 {
		toxColor = core.Red
	} else if survival.Toxicity > 20.0 {
		toxColor = core.Yellow
	}
	e.drawText(70, hudY+1, fmt.Sprintf("TOX [%s] %3.0f%%", toxBar, survival.Toxicity), toxColor)
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
	e.drawText(clockX, hudY+2, clockText, core.Yellow)
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
		if i == 0 && len(e.Messages) == 3 {
			color = display.DarkenColor(core.Green, 3)
		}
		if i == 1 && len(e.Messages) == 3 {
			color = display.DarkenColor(core.Green, 1)
		}
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
	var interactPrompt string
	var survival components.PlayerSurvival

	if found {
		statusText = ctrl.Status.Title()
		interactPrompt = e.getNearbyInteractionPrompt(pos.X, pos.Y)
		survival = ctrl.Survival
	}

	e.drawHUDStatusAndNav(hudY, statusText, ctrl.Autopilot, survival)
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

