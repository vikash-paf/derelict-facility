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
}

func NewEngine(term *terminal.Terminal, width, height int) *Engine {
	return &Engine{
		Terminal:   term,
		Map:        world.NewMap(width, height),
		Player:     entity.NewActor(width/2, height/2, '@'),
		Running:    true,
		TickerRate: time.Millisecond * 33, // ~30 fps
	}
}

// Run starts the deterministic game loop
func (e *Engine) Run() error {
	inputChan := e.Terminal.PollInput()
	ticker := time.NewTicker(e.TickerRate)
	defer ticker.Stop()

	width, height := e.Map.Width, e.Map.Height
	memorySize := (width * height) + (height * 2) + 50 // bytes, a little bit extra memory
	screen := bytes.NewBuffer(make([]byte, 0, memorySize))

	for e.Running {
		screen.Reset()

		// move the cursor to the top left
		screen.WriteString("\033[H")

		// spawn the player
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if e.Player.X == x && e.Player.Y == y {
					screen.WriteRune(e.Player.Char)
				} else {
					screen.WriteString(".")
				}
			}

			screen.WriteString("\r\n") // In raw mode, \n just moves down, \r moves to start of line
		}

		os.Stdout.Write(screen.Bytes())

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
	switch event.Key {
	case 'w':
		e.Player.Y--
	case 's':
		e.Player.Y++
	case 'a':
		e.Player.X--
	case 'd':
		e.Player.X++
	case 'q':
		e.Running = false
	}
}
