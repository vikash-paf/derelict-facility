package mission

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/vikash-paf/derelict-facility/assets"
)

// LevelMeta describes a single level in a campaign mission.
type LevelMeta struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	File string `json:"file"`
}

// MissionManifest defines the storyline metadata for a mission campaign package.
type MissionManifest struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Author     string      `json:"author"`
	Synopsis   string      `json:"synopsis"`
	StartLevel string      `json:"start_level"`
	Levels     []LevelMeta `json:"levels"`

	// Directory path where mission files reside (e.g. "missions/sector_4_incident")
	Dir string `json:"-"`
}

// MissionLoader loads story campaign manifests from embedded or external files.
type MissionLoader struct{}

func NewMissionLoader() *MissionLoader {
	return &MissionLoader{}
}

// DiscoverMissions scans embedded assets/missions for valid mission package manifests.
func (l *MissionLoader) DiscoverMissions() ([]MissionManifest, error) {
	var manifests []MissionManifest

	entries, err := assets.AssetsFS.ReadDir("missions")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded missions directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		manifestPath := fmt.Sprintf("missions/%s/mission.json", dirName)

		data, err := assets.AssetsFS.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var m MissionManifest
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}

		m.Dir = fmt.Sprintf("missions/%s", dirName)
		manifests = append(manifests, m)
	}

	return manifests, nil
}

// LoadLevelMapData retrieves the raw JSON byte data for a level map file inside a mission.
func (m *MissionManifest) LoadLevelMapData(fileRelativePath string) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/%s", m.Dir, fileRelativePath)

	// Try reading from embedded filesystem first
	data, err := assets.AssetsFS.ReadFile(fullPath)
	if err == nil {
		return data, nil
	}

	// Fallback: try direct OS filesystem read
	return fs.ReadFile(osFS{}, fullPath)
}

type osFS struct{}

func (osFS) Open(name string) (fs.File, error) {
	return nil, fs.ErrNotExist
}
