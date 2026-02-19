package main

import (
	"fmt"
	"strings"
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

	var screen strings.Builder

	running := true

	for running {
		screen.Reset()

		// move the cursor to the top left
		screen.WriteString("\033[H")

		// spawn the player
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if player.X == x && player.Y == y {
					screen.WriteRune(player.Char)
				} else {
					screen.WriteString(".")
				}
			}

			screen.WriteString("\r\n") // In raw mode, \n just moves down, \r moves to start of line
		}

		fmt.Print(screen.String())

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
