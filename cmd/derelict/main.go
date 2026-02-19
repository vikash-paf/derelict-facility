package main

import (
	"fmt"

	"github.com/vikash-paf/derelict-facility/internal/engine"
	"github.com/vikash-paf/derelict-facility/internal/terminal"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func main() {
	width, height := 80, 25
	gameMap := world.NewMap(width, height)

	for i := 0; i < width*height; i++ {
		gameMap.Tiles[i] = world.Tile{Type: world.TileTypeFloor, Walkable: true}
	}

	term := terminal.NewTerminal()
	err := term.EnableRawMode()
	if err != nil {
		panic(err)
	}
	defer term.Restore()

	gameEngine := engine.NewEngine(term, width, height)

	// Draw a simple box of walls around the edges
	for x := 0; x < 80; x++ {
		gameEngine.Map.SetTile(x, 0, world.Tile{Type: world.TileTypeWall, Walkable: false})
		gameEngine.Map.SetTile(x, 23, world.Tile{Type: world.TileTypeWall, Walkable: false})
	}

	err = gameEngine.Run()
	if err != nil {
		fmt.Errorf(err.Error())
	}
}
