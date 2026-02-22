package engine

import (
	"bytes"
	"math/rand"
	"os"
	"time"

	"github.com/vikash-paf/derelict-facility/internal/entity"
	"github.com/vikash-paf/derelict-facility/internal/terminal"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

const (
	cursorHome = "\033[H"
	lineBreak  = "\r\n"
	fovRadius  = 8 // cool stuff can be done here, like a dimming torch light
)

type Engine struct {
	Terminal   *terminal.Terminal
	Map        *world.Map
	Player     *world.Player
	Theme      world.TileVariant
	screen     bytes.Buffer
	TickerRate time.Duration
	tickCount  int
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
	case 'p':
		e.Player.Autopilot = !e.Player.Autopilot
		e.Player.CurrentPath = nil // Clear the path if we toggle it
	case 'q':
		e.Running = false
		return
	}

	if dx != 0 || dy != 0 {
		e.movePlayer(dx, dy)
	}
}

func (e *Engine) movePlayer(dx, dy int) {
	newX := e.Player.X + dx
	newY := e.Player.Y + dy

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

	pathLookup := make(map[int]bool)
	if e.Player.Autopilot {
		for _, p := range e.Player.CurrentPath {
			pathLookup[p.Y*e.Map.Width+p.X] = true
		}
	}

	for y := 0; y < e.Map.Height; y++ {
		for x := 0; x < e.Map.Width; x++ {
			// 1. Render the player
			if e.Player.X == x && e.Player.Y == y {
				e.screen.WriteString(e.Player.Render())
				continue
			}

			tile := e.Map.GetTile(x, y)
			if tile == nil {
				continue
			}
			isPathTile := pathLookup[y*e.Map.Width+x]
			// We only draw the path if it's on a tile we've at least explored!
			// (Drawing a path through Pitch Black space breaks the Fog of War illusion).
			if isPathTile && (tile.Visible || tile.Explored) {
				// Use a dim character like a period or asterisk.
				// If you have a Yellow or Red ANSI code, use it here to make it look like a scanner!
				e.screen.WriteString(world.Red + "*" + world.Reset)
				continue
			}

			// 2. Render the map tiles
			if tile.Visible {
				e.screen.WriteString(e.Theme[tile.Type])
				continue
			}

			if tile.Explored {
				// We wrap the character in Gray and Reset to "dim" the lights
				// Note: We use the character from Classic to keep the 'memory' simple
				char := e.Theme[tile.Type]
				e.screen.WriteString(world.Gray + char + world.Reset)
				continue
			}

			e.screen.WriteString(e.Theme[world.TileTypeEmpty])
		}

		e.screen.WriteString(lineBreak)
	}

	// Render the screen
	os.Stdout.Write(e.screen.Bytes())
}

func (e *Engine) Update() {
	e.tickCount++

	// Run AI movement every 2nd frame (approx 15 times a second)
	if e.Player.Autopilot && e.tickCount%2 == 0 {
		// run autopilot
		e.processAutopilot()
	}

	e.Map.ComputeFOV(e.Player.X, e.Player.Y, fovRadius)

	// this is where I will add:
	// - Check if the player stepped on a checkpoint
	// - Move enemies
	// - Trigger story events
	// - Update flashing light animations
}

func (e *Engine) processAutopilot() {
	// 1. If we don't have a path, find a new destination!
	if len(e.Player.CurrentPath) == 0 {
		// Pick a random room
		targetRoom := e.Map.Rooms[rand.Intn(len(e.Map.Rooms))]
		targetX, targetY := targetRoom.Center()

		start := entity.Point{X: e.Player.X, Y: e.Player.Y}
		target := entity.Point{X: targetX, Y: targetY}

		// Calculate the path
		path := world.FindPath(e.Map, start, target)

		// A* returns the starting node as index 0. We slice it off so we don't stand still.
		if len(path) > 1 {
			e.Player.CurrentPath = path[1:]
		} else {
			e.Player.CurrentPath = nil // We are already there, or trapped
		}

		return
	}

	// 2. Take the next step in the path
	nextStep := e.Player.CurrentPath[0]

	// Ensure the tile is still walkable (in case a door closed or enemy moved there)
	if e.Map.IsWalkable(nextStep.X, nextStep.Y) {
		e.Player.X = nextStep.X
		e.Player.Y = nextStep.Y
	} else {
		// Path is blocked! Clear it so we recalculate next tick.
		e.Player.CurrentPath = nil
		return
	}

	// 3. Pop the step we just took off the slice
	e.Player.CurrentPath = e.Player.CurrentPath[1:]
}
