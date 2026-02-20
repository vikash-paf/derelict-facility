package engine

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/terminal"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

const (
	cursorHome = "\033[H"
	lineBreak  = "\r\n"
)

type Engine struct {
	Terminal *terminal.Terminal
	Map      *world.Map
	Player   *entity.Actor

	Theme world.TileVariant

	Running    bool
	TickerRate time.Duration
	screen     bytes.Buffer
}

func NewEngine(
	term *terminal.Terminal,
	width, height int,
	startingTheme world.TileVariant,
) *Engine {
	e := &Engine{
		Terminal:   term,
		Map:        world.NewMap(width, height),
		Running:    true,
		TickerRate: time.Millisecond * 33, // ~30 fps
		Theme:      startingTheme,
	}

	memorySize := (width * height) + (height * 2) + 50 // bytes, a little bit extra memory
	e.screen.Grow(memorySize)

	e.Player = entity.NewActor(width/2, height/2, '@')

	return e
}

// Run starts the deterministic game loop
func (e *Engine) Run() error {
	inputChan := e.Terminal.PollInput()

	ticker := time.NewTicker(e.TickerRate)
	defer ticker.Stop()

	for e.Running {
		select {
		case event := <-inputChan:
			e.handleInput(event)
		case <-ticker.C:
			// todo: other logic here, like other characters moving around
			e.render()
		}
	}

	return nil
}

func (e *Engine) handleInput(event terminal.InputEvent) {
	dx, dy := 0, 0

	switch event.Key {
	case 'w':
		dy = -1
	case 's':
		dy = 1
	case 'a':
		dx = -1
	case 'd':
		dx = 1
	case 'q':
		e.Running = false
	}

	if dx != 0 || dy != 0 {
		e.movePlayer(dx, dy)
	}
}

func (e *Engine) movePlayer(x, y int) {
	newX := e.Player.X + x
	newY := e.Player.Y + y

	// if outside the map, do nothing
	if newX < 0 || newX >= e.Map.Width || newY < 0 || newY >= e.Map.Height {
		return
	}

	// check if the tile is walkable
	tile := e.Map.GetTile(newX, newY)
	if tile != nil && tile.Walkable {
		e.Player.X = newX
		e.Player.Y = newY
	}
}

func (e *Engine) render() {
	e.screen.Reset()
	e.screen.WriteString(cursorHome)

	for y := 0; y < e.Map.Height; y++ {
		for x := 0; x < e.Map.Width; x++ {
			// 1. Render the player
			if e.Player.X == x && e.Player.Y == y {
				e.screen.WriteRune(e.Player.Char)
				continue
			}

			// 2. Render the map tiles
			tile := e.Map.GetTile(x, y)
			if tile == nil {
				continue
			}

			switch tile.Type {

			// v0
			// case world.TileTypeWall:
			// 	e.screen.WriteString("#")
			// case world.TileTypeFloor:
			// 	e.screen.WriteString("·")
			// case world.TileTypeEmpty:
			// 	e.screen.WriteString(" ")

			// v1
			case world.TileTypeWall:
				e.screen.WriteString("█")
			case world.TileTypeFloor:
				e.screen.WriteString("·")
			case world.TileTypeEmpty:
				e.screen.WriteString(" ")

			// v2
			// case world.TileTypeWall:
			// 	e.screen.WriteString("▓")
			// case world.TileTypeFloor:
			// 	e.screen.WriteString("░")
			// case world.TileTypeEmpty:
			// 	e.screen.WriteString(" ")

			// v3
			// case world.TileTypeWall:
			// 	e.screen.WriteString("■")
			// case world.TileTypeFloor:
			// 	e.screen.WriteString(".")
			// case world.TileTypeEmpty:
			// 	e.screen.WriteString(" ")

			default:
				panic(fmt.Sprintf("unknown tile type: %d", tile.Type))
			}
		}

		e.screen.WriteString(lineBreak)
	}

	// Render the screen
	os.Stdout.Write(e.screen.Bytes())
}
