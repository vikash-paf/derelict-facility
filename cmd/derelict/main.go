package main

import (
	"fmt"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/engine"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func main() {
	mapWidth, mapHeight := 129, 40
	windowWidth, windowHeight := 120, 45

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

	// 2. Build the world map FIRST
	seed := time.Now().UnixNano()
	generator := world.NewFacilityGenerator(uint64(seed))
	generatedMap, playerX, playerY := generator.Generate(mapWidth, mapHeight)
	if generatedMap == nil {
		panic("Failed to generate map")
	}

	// 3. Setup the ECS and spawn the Player
	ecsWorld := ecs.NewWorld()

	playerEnt := ecsWorld.CreateEntity()
	ecsWorld.AddPosition(playerEnt, components.Position{X: playerX, Y: playerY})
	ecsWorld.AddGlyph(playerEnt, components.Glyph{Char: "@", ColorCode: world.BrightWhite}) // Astronaut
	ecsWorld.AddPlayerControl(playerEnt, components.PlayerControl{
		Autopilot: false,
		Status:    components.PlayerStatusHealthy,
	})

	// 5. Spawn a test Power Generator
	genEnt := ecsWorld.CreateEntity()
	ecsWorld.AddPosition(genEnt, components.Position{X: playerX + 2, Y: playerY})
	ecsWorld.AddGlyph(genEnt, components.Glyph{Char: "X", ColorCode: world.Red})
	ecsWorld.AddSolid(genEnt)
	ecsWorld.AddInteractable(genEnt, components.Interactable{Prompt: "Press [E] to Toggle Generator"})
	ecsWorld.AddPowerGenerator(genEnt, components.PowerGenerator{IsActive: false})

	// 6. Spawn Doors
	for _, doorPos := range generatedMap.Doors {
		// Don't spawn a door right on top of the player or generator
		if doorPos.X == playerX && doorPos.Y == playerY {
			continue
		}
		if doorPos.X == playerX+2 && doorPos.Y == playerY {
			continue
		}

		doorEnt := ecsWorld.CreateEntity()
		ecsWorld.AddPosition(doorEnt, components.Position{X: doorPos.X, Y: doorPos.Y})
		ecsWorld.AddGlyph(doorEnt, components.Glyph{Char: "+", ColorCode: world.White})
		ecsWorld.AddSolid(doorEnt) // Closed doors block movement!
		ecsWorld.AddInteractable(doorEnt, components.Interactable{Prompt: "Press [E] to Open Door"})
		ecsWorld.AddDoor(doorEnt, components.Door{IsOpen: false})
	}

	// 7. Hand everything to the Engine
	gameEngine := engine.NewEngine(disp, generatedMap, ecsWorld, world.TileVariantGritty)

	err = gameEngine.Run()
	if err != nil {
		fmt.Println(err)
	}
}
