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
}

func saveState(w *ecs.World, m *world.Map, activeMission *mission.MissionManifest, activeLevelID string) {
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
