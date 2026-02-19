package engine

import (
	"bytes"
	"os"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/terminal"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

type Engine struct {
	Terminal *terminal.Terminal
	Map      *world.Map
	Player   *entity.Actor

	Running    bool
	TickerRate time.Duration
	screen     bytes.Buffer
}

func NewEngine(term *terminal.Terminal, width, height int) *Engine {
	e := &Engine{
		Terminal:   term,
		Map:        world.NewMap(width, height),
		Running:    true,
		TickerRate: time.Millisecond * 33, // ~30 fps
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

	width, height := e.Map.Width, e.Map.Height

	for e.Running {
		e.screen.Reset()

		// move the cursor to the top left
		e.screen.WriteString("\033[H")

		// spawn the player
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if e.Player.X == x && e.Player.Y == y {
					e.screen.WriteRune(e.Player.Char)
				} else {
					e.screen.WriteString(".")
				}
			}

			e.screen.WriteString("\r\n") // In raw mode, \n just moves down, \r moves to start of line
		}

		os.Stdout.Write(e.screen.Bytes())

		// handle input and move player
		select {
		case event := <-inputChan:
			e.handleInput(event)
		default:
			// todo: add fps limit (sleep here)
		}

		time.Sleep(16 * time.Millisecond) // ~60FPS
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
