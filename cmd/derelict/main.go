package main

import (
	"fmt"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/display"
	"github.com/vikash-paf/derelict-facility/internal/engine"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func main() {
	width, height := 120, 26 // 120x30 -- reserving 4 rows for the HUD
	tileSize := int32(16)    // From the handover doc
	fontSize := int32(20)    // Optional: make font size slightly larger for readability

	disp := display.NewRaylibDisplay(tileSize, fontSize)

	err := disp.Init(width, height, "Derelict Facility")
	if err != nil {
		panic(err)
	}
	defer disp.Close()

	// 2. Build the world and player FIRST
	seed := time.Now().UnixNano()
	generator := world.NewFacilityGenerator(uint64(seed))
	generatedMap, playerX, playerY := generator.Generate(width, height)

	player := world.NewPlayer(playerX, playerY, world.PlayerStatusHealthy, false)

	// 3. Hand them to the Engine
	gameEngine := engine.NewEngine(disp, generatedMap, player, world.TileVariantCold)

	err = gameEngine.Run()
	if err != nil {
		fmt.Println(err)
	}
}
