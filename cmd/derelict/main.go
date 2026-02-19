package main

import (
	"fmt"
	"time"

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

	inputChan := term.PollInput()

	running := true

	for running {
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

		// handle input and move player
		select {
		case event := <-inputChan:
			switch event.Key {
			case 'w':
				player.Y--
			case 's':
				player.Y++
			case 'a':
				player.X--
			case 'd':
				player.X++
			case 'q':
				running = false
			}
		default:
			// todo: add fps limit (sleep here)
		}

		time.Sleep(16 * time.Millisecond) // ~60FPS
	}
}
