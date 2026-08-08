package world

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONMapLoader(t *testing.T) {
	jsonContent := `{
		"width": 10,
		"height": 5,
		"rows": [
			"##########",
			"#@...T...#",
			"#.####+###",
			"#..G.....#",
			"##########"
		]
	}`

	tmpDir := t.TempDir()
	mapPath := filepath.Join(tmpDir, "test_map.json")
	if err := os.WriteFile(mapPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to create temp map file: %v", err)
	}

	loader := NewJSONMapLoader()
	m, playerX, playerY, err := loader.Load(mapPath)
	if err != nil {
		t.Fatalf("unexpected error loading map: %v", err)
	}

	if m == nil {
		t.Fatalf("expected map to be non-nil")
	}
	if m.Width != 10 || m.Height != 5 {
		t.Errorf("expected map dimensions 10x5, got %dx%d", m.Width, m.Height)
	}
	if playerX != 1 || playerY != 1 {
		t.Errorf("expected player pos (1,1), got (%d,%d)", playerX, playerY)
	}
	if len(m.Doors) != 1 {
		t.Errorf("expected 1 door, got %d", len(m.Doors))
	}
	if len(m.Terminals) != 1 {
		t.Errorf("expected 1 terminal, got %d", len(m.Terminals))
	}
	if len(m.PowerGenerators) != 1 {
		t.Errorf("expected 1 power generator, got %d", len(m.PowerGenerators))
	}
}
