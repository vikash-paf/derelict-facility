package engine

import (
	"bytes"
	"os"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/terminal"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

const (
	cursorHome = "\033[H"
	lineBreak  = "\r\n"
)

type Engine struct {
	Terminal   *terminal.Terminal
	Map        *world.Map
	Player     *world.Player
	Theme      world.TileVariant
	screen     bytes.Buffer
	TickerRate time.Duration
	Running    bool
}

func NewEngine(
	term *terminal.Terminal,
	gameMap *world.Map,
	player *world.Player,
	startingTheme world.TileVariant,
) *Engine {
	e := &Engine{
		Terminal:   term,
		Map:        gameMap,
		Player:     player,
		Running:    true,
		TickerRate: time.Millisecond * 33, // ~30 fps
		Theme:      startingTheme,
	}

	// Calculate memory based on the map handed to us
	memorySize := (gameMap.Width * gameMap.Height) + (gameMap.Height * 2) + 50
	e.screen.Grow(memorySize)

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
			e.Update() // Calculate all game rules!
			e.render() // Paint the results!
		}
	}

	return nil
}

func (e *Engine) handleInput(event terminal.InputEvent) {
	dx, dy := 0, 0

	if event.Quit {
		e.Running = false
		return
	}

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
		return
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
				e.screen.WriteString(e.Player.Render())
				continue
			}

			// 2. Render the map tiles
			tile := e.Map.GetTile(x, y)
			if tile == nil {
				continue
			}

			e.screen.WriteString(e.Theme[tile.Type])
		}

		e.screen.WriteString(lineBreak)
	}

	// Render the screen
	os.Stdout.Write(e.screen.Bytes())
}

func (e *Engine) Update() {
	// Right now, this is empty!
	// But tomorrow, this is where I will add:
	// - Check if the player stepped on a checkpoint
	// - Move enemies
	// - Trigger story events
	// - Update flashing light animations
}
