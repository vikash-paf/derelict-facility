package main

import (
	"fmt"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/engine"
	"github.com/vikash-paf/derelict-facility/internal/terminal"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func main() {
	width, height := 80, 25

	term := terminal.NewTerminal()
	err := term.EnableRawMode()
	if err != nil {
		panic(err)
	}
	defer term.Restore()

	gameEngine := engine.NewEngine(term, width, height, world.TileVariantCold)

	seed := time.Now().UnixNano()
	generator := world.NewFacilityGenerator(uint64(seed))

	generatedMap, playerX, playerY := generator.Generate(width, height)
	gameEngine.Map = generatedMap
	gameEngine.Player.X = playerX
	gameEngine.Player.Y = playerY

	err = gameEngine.Run()
	if err != nil {
		fmt.Println(err)
	}
}
