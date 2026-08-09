package main

import (
	"flag"
	"fmt"

	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/engine"
	"github.com/vikash-paf/derelict-facility/internal/systems"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func main() {
	flag.Parse()

	viewWidth, viewHeight := 100, 30 // Viewport grid size
	hudHeight := 8                   // Room for message log HUD
	windowWidth, windowHeight := viewWidth, viewHeight+hudHeight

	cellWidth := int32(20)
	cellHeight := int32(40)
	fontSize := int32(40)
	fontPath := "fonts/FiraCodeNFBoldMono.ttf"

	disp := display.NewRaylibDisplay(cellWidth, cellHeight, fontSize, fontPath)

	if err := disp.Init(windowWidth, windowHeight, "Derelict Facility"); err != nil {
		panic(err)
	}
	defer disp.Close()

	// Try to load saved game state first
	saveData, err := systems.LoadState()
	var loadedMap *world.Map
	var ecsWorld *ecs.World

	if err == nil && saveData != nil {
		fmt.Println("Loaded savegame.sav")
		loadedMap = saveData.Map
		ecsWorld = saveData.World
	}

	gameEngine := engine.NewEngine(
		disp,
		loadedMap,
		ecsWorld,
		world.TileVariantGritty,
		viewWidth,
		viewHeight,
	)

	// If loaded from savegame, skip main menu directly into gameplay
	if saveData != nil {
		gameEngine.ActiveMission = saveData.ActiveMission
		gameEngine.ActiveLevelID = saveData.ActiveLevelID
		gameEngine.State = engine.GameStateRunning
	}

	if err := gameEngine.Run(); err != nil {
		panic(err)
	}
}
