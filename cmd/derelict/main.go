package main

import (
	"flag"
	"fmt"

	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/core"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/engine"
	"github.com/vikash-paf/derelict-facility/internal/systems"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func main() {
	mapFile := flag.String("map", "", "path to a JSON map file to load")
	flag.Parse()

	mapWidth, mapHeight := 120, 40
	viewWidth, viewHeight := 80, 25 // Camera sees 80x25 tiles
	hudHeight := 6
	windowWidth, windowHeight := viewWidth, viewHeight+hudHeight

	cellWidth := int32(10)
	cellHeight := int32(20)
	fontSize := int32(20)
	fontPath := "assets/fonts/FiraCodeNFBoldMono.ttf"

	disp := display.NewRaylibDisplay(cellWidth, cellHeight, fontSize, fontPath)

	err := disp.Init(windowWidth, windowHeight, "Derelict Facility")
	if err != nil {
		panic(err)
	}
	defer disp.Close()

	// Try to load savegame first
	saveData, err := systems.LoadState()
	var generatedMap *world.Map
	var ecsWorld *ecs.World

	if err == nil && saveData != nil {
		fmt.Println("Loaded savegame.sav")
		generatedMap = saveData.Map
		ecsWorld = saveData.World
	} else {
		fmt.Println("No savegame found, generating new world...")
		var playerX, playerY int
		if *mapFile != "" {
			fmt.Println("Loading map from JSON:", *mapFile)
			loader := world.NewJSONMapLoader()
			var err error
			generatedMap, playerX, playerY, err = loader.Load(*mapFile)
			if err != nil {
				panic(fmt.Sprintf("Failed to load map from %s: %v", *mapFile, err))
			}
		} else {
			// 2. Build the world map FIRST
			// seed := time.Now().UnixNano()
			seed := 12345
			generator := world.NewFacilityGenerator(uint64(seed))
			generatedMap, playerX, playerY = generator.Generate(mapWidth, mapHeight)
			if generatedMap == nil {
				panic("Failed to generate map")
			}
		}

		// 3. Setup the ECS and spawn the Player
		ecsWorld = ecs.NewWorld()

		playerEnt := ecsWorld.CreateEntity()
		ecsWorld.AddPosition(playerEnt, components.Position{X: playerX, Y: playerY})
		ecsWorld.AddGlyph(playerEnt, components.Glyph{Char: "@", Color: core.BrightWhite}) // Astronaut
		ecsWorld.AddPlayerControl(playerEnt, components.PlayerControl{
			Autopilot: false,
			Status:    components.PlayerStatusHealthy,
		})

		// 5. Spawn Power Generators from Map
		for _, genPos := range generatedMap.PowerGenerators {
			genEnt := ecsWorld.CreateEntity()
			ecsWorld.AddPosition(genEnt, components.Position{X: genPos.X, Y: genPos.Y})
			ecsWorld.AddGlyph(genEnt, components.Glyph{Char: "X", Color: core.Red})
			ecsWorld.AddSolid(genEnt)
			ecsWorld.AddInteractable(genEnt, components.Interactable{Prompt: "Press [E] to Toggle Generator"})
			ecsWorld.AddPowerGenerator(genEnt, components.PowerGenerator{IsActive: false})
		}

		// Spawn Terminals from Map
		for _, termPos := range generatedMap.Terminals {
			termEnt := ecsWorld.CreateEntity()
			ecsWorld.AddPosition(termEnt, components.Position{X: termPos.X, Y: termPos.Y})
			ecsWorld.AddGlyph(termEnt, components.Glyph{Char: "T", Color: core.Cyan})
			ecsWorld.AddSolid(termEnt)
			ecsWorld.AddInteractable(termEnt, components.Interactable{Prompt: "Press [E] to Access Terminal"})
			ecsWorld.AddTerminal(termEnt, components.Terminal{HasSaved: false})
			ecsWorld.AddNarrative(termEnt, components.Narrative{Text: "LOG 001: Sector 4 containment breach. All non-essential personnel evacuate immediately. The life support backup is failing."})
		}

		// 6. Spawn Doors
		for _, doorPos := range generatedMap.Doors {
			// Don't spawn a door right on top of the player
			if doorPos.X == playerX && doorPos.Y == playerY {
				continue
			}

			doorEnt := ecsWorld.CreateEntity()
			ecsWorld.AddPosition(doorEnt, components.Position{X: doorPos.X, Y: doorPos.Y})
			ecsWorld.AddGlyph(doorEnt, components.Glyph{Char: "+", Color: core.White})
			ecsWorld.AddSolid(doorEnt) // Closed doors block movement!
			ecsWorld.AddInteractable(doorEnt, components.Interactable{Prompt: "Press [E] to Open Door"})
			ecsWorld.AddDoor(doorEnt, components.Door{IsOpen: false})
		}
	}

	// 7. Hand everything to the Engine
	gameEngine := engine.NewEngine(
		disp,
		generatedMap,
		ecsWorld,
		world.TileVariantGritty,
		viewWidth,
		viewHeight,
	)

	err = gameEngine.Run()
	if err != nil {
		fmt.Println(err)
	}
}
