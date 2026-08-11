package systems

import (
	"encoding/gob"
	"os"

	"github.com/vikash-paf/derelict-facility/internal/ecs"
	"github.com/vikash-paf/derelict-facility/internal/mission"
	"github.com/vikash-paf/derelict-facility/internal/world"
)

type SaveGameData struct {
	World         *ecs.World
	Map           *world.Map
	ActiveMission *mission.MissionManifest
	ActiveLevelID string
	// Save persistent caches
	TotalTicks uint64
	Day        int
	Season     uint8
}

func SaveState(w *ecs.World, m *world.Map, activeMission *mission.MissionManifest, activeLevelID string, ticks uint64, day int, season uint8) {
	file, err := os.Create("savegame.sav")
	if err != nil {
		return
	}
	defer file.Close()

	data := SaveGameData{
		World:         w,
		Map:           m,
		ActiveMission: activeMission,
		ActiveLevelID: activeLevelID,
		TotalTicks:    ticks,
		Day:           day,
		Season:        season,
	}

	encoder := gob.NewEncoder(file)
	encoder.Encode(data)
}

func LoadState() (*SaveGameData, error) {
	file, err := os.Open("savegame.sav")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data SaveGameData
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
