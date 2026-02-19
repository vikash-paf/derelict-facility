package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/terminal"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func main() {
	width, height := 80, 25
	gameMap := world.NewMap(width, height)

	for i := 0; i < width*height; i++ {
		gameMap.Tiles[i] = world.Tile{Type: world.TileTypeFloor, Walkable: true}
	}

	// create the player
	player := entity.NewActor(10, 10, '@')

	term := terminal.NewTerminal()
	err := term.EnableRawMode()
	if err != nil {
		panic(err)
	}
	defer term.Restore()

	term.Clear()

	// spawn the player
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if player.X == x && player.Y == y {
				fmt.Printf("%c", player.Char)
			} else {
				fmt.Printf(".")
			}
		}

		fmt.Print("\r\n") // In raw mode, \n just moves down, \r moves to start of line
	}

	// read the input to verify if the raw mode works, 'q' to quit
	reader := bufio.NewReader(os.Stdin)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return
		}

		if r == 'q' {
			break
		}
	}

}
