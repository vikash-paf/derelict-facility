package main

import (
	"fmt"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/engine"
	"github.com/vikash-paf/derelict-facility/internal/terminal"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func main() {
	width, height := 120, 26 // 120x30 -- reserving 4 rows for the HUD

	term := terminal.NewTerminal()
	err := term.EnableRawMode()
	if err != nil {
		panic(err)
	}
	defer term.Restore()

	// 1. Get the terminal size
	// todo: (there is a sizing issue with my terminal emulator, so use the 120x30)
	_, _, err = term.GetSize()
	if err != nil {
		panic(err)
	}

	// 2. Build the world and player FIRST
	seed := time.Now().UnixNano()
	generator := world.NewFacilityGenerator(uint64(seed))
	generatedMap, playerX, playerY := generator.Generate(width, height)

	player := world.NewPlayer(playerX, playerY, world.PlayerStatusHealthy, false)

	// 3. Hand them to the Engine
	gameEngine := engine.NewEngine(term, generatedMap, player, world.TileVariantCold)

	err = gameEngine.Run()
	if err != nil {
		fmt.Println(err)
	}
}
