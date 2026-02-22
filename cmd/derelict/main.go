package main

import (
	"fmt"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/engine"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func main() {
	mapWidth, mapHeight := 120, 26
	windowWidth, windowHeight := 120, 30

	cellWidth := int32(12)
	cellHeight := int32(24)
	fontSize := int32(24)
	fontPath := "assets/fonts/FiraCode-Bold.ttf"

	disp := display.NewRaylibDisplay(cellWidth, cellHeight, fontSize, fontPath)

	err := disp.Init(windowWidth, windowHeight, "Derelict Facility")
	if err != nil {
		panic(err)
	}
	defer disp.Close()

	// 2. Build the world and player FIRST
	seed := time.Now().UnixNano()
	generator := world.NewFacilityGenerator(uint64(seed))
	generatedMap, playerX, playerY := generator.Generate(mapWidth, mapHeight)

	player := world.NewPlayer(playerX, playerY, world.PlayerStatusHealthy, false)

	// 3. Hand them to the Engine
	gameEngine := engine.NewEngine(disp, generatedMap, player, world.TileVariantCold)

	err = gameEngine.Run()
	if err != nil {
		fmt.Println(err)
	}
}
