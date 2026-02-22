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
	mapWidth, mapHeight := 20, 12
	windowWidth, windowHeight := 20, 15

	cellWidth := int32(64)
	cellHeight := int32(64)
	fontSize := int32(18)
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

	// 3. Setup the ECS and spawn the Player
	ecsWorld := ecs.NewWorld()

	playerEnt := ecsWorld.CreateEntity()
	ecsWorld.AddComponent(playerEnt, components.NamePosition, &components.Position{X: playerX, Y: playerY})
	ecsWorld.AddComponent(playerEnt, components.NameSprite, &components.Sprite{SheetX: 0, SheetY: 6, ColorCode: world.White}) // Front-facing engineer (Row 6)
	ecsWorld.AddComponent(playerEnt, components.NamePlayerControl, &components.PlayerControl{
		Autopilot: false,
		Status:    components.PlayerStatusHealthy,
	})

	// 4. Hand everything to the Engine
	gameEngine := engine.NewEngine(disp, generatedMap, ecsWorld, world.TileVariantGritty)

	err = gameEngine.Run()
	if err != nil {
		fmt.Println(err)
	}
}
