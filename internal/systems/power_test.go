package systems

import (
	"testing"

	"github.com/vikash-paf/derelict-facility/internal/components"
	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

func TestPowerGrid_GlobalVsLocal(t *testing.T) {
	w := ecs.NewWorld()
	gameMap := world.NewMap(50, 50)
	gameMap.Rooms = []world.Rect{
		{X1: 0, Y1: 0, X2: 10, Y2: 10},  // Room 0
		{X1: 20, Y1: 20, X2: 30, Y2: 30}, // Room 1
	}

	// Test 1: No active generators -> power inactive everywhere
	if IsPowerActive(w) {
		t.Errorf("expected global power to be inactive initially")
	}
	if IsPowerActiveAt(w, gameMap, 5, 5) {
		t.Errorf("expected room 0 to be unpowered initially")
	}

	// Test 2: Local Generator in Room 0
	localGen := w.CreateEntity()
	w.AddPosition(localGen, components.Position{X: 2, Y: 2})
	w.AddPowerGenerator(localGen, components.PowerGenerator{IsActive: true, IsGlobal: false})

	// Room 0 should be powered, but global power and Room 1 should not
	if IsPowerActive(w) {
		t.Errorf("expected IsPowerActive to be false for local generator")
	}
	if !IsPowerActiveAt(w, gameMap, 5, 5) {
		t.Errorf("expected room 0 (5,5) to be powered by local generator at (2,2)")
	}
	if IsPowerActiveAt(w, gameMap, 25, 25) {
		t.Errorf("expected room 1 (25,25) to remain unpowered when local generator is in room 0")
	}

	// Test 3: Global Generator activated
	globalGen := w.CreateEntity()
	w.AddPosition(globalGen, components.Position{X: 22, Y: 22})
	w.AddPowerGenerator(globalGen, components.PowerGenerator{IsActive: true, IsGlobal: true})

	// Global power should now power everywhere
	if !IsPowerActive(w) {
		t.Errorf("expected IsPowerActive to be true for global generator")
	}
	if !IsPowerActiveAt(w, gameMap, 25, 25) {
		t.Errorf("expected room 1 (25,25) to be powered by global generator")
	}
}
