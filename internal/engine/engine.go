package engine

import (
	"fmt"
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
	// TODO: implement game loop
	fmt.Println("Engine running...")
	return nil
}
