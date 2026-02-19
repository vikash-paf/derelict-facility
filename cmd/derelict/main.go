package main

import (
	"fmt"

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

	fmt.Println("Starting derelict-facility engine...")
}
