package engine

import (
	"fmt"

	"github.com/vikash-paf/derelict-facility/internal/terminal"
)

// Engine manages the game loop and time management
type Engine struct {
	term *terminal.Terminal
	// Add other dependencies: world, entities
}

// New creates a new engine instance
func New(term *terminal.Terminal) *Engine {
	return &Engine{
		term: term,
	}
}

// Run starts the deterministic game loop
func (e *Engine) Run() error {
	// TODO: implement game loop
	fmt.Println("Engine running...")
	return nil
}
