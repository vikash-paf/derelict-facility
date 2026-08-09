package mission

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/vikash-paf/derelict-facility/assets"
)

var missionTitleRegex = regexp.MustCompile(`^(?i)MISSION\s+\S+\s*:`)

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

// DiscoverMissions scans embedded assets/missions, local OS assets/missions, and local OS `./missions` directory, merging and indexing them correctly.
func (l *MissionLoader) DiscoverMissions() ([]MissionManifest, error) {
	manifestMap := make(map[string]MissionManifest)

	// 1. Scan embedded assets first
	if entries, err := assets.AssetsFS.ReadDir("missions"); err == nil {
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
			manifestMap[m.ID] = m
		}
	}

	// 2. Scan local OS assets/missions directory (useful for non-compiled / live dev mode)
	if entries, err := os.ReadDir("assets/missions"); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dirName := entry.Name()
			manifestPath := fmt.Sprintf("assets/missions/%s/mission.json", dirName)

			data, err := os.ReadFile(manifestPath)
			if err != nil {
				continue
			}

			var m MissionManifest
			if err := json.Unmarshal(data, &m); err != nil {
				continue
			}

			m.Dir = fmt.Sprintf("assets/missions/%s", dirName)
			manifestMap[m.ID] = m
		}
	}

	// 3. Scan local OS ./missions directory (parallel to the binary file)
	if entries, err := os.ReadDir("missions"); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dirName := entry.Name()
			manifestPath := fmt.Sprintf("missions/%s/mission.json", dirName)

			data, err := os.ReadFile(manifestPath)
			if err != nil {
				continue
			}

			var m MissionManifest
			if err := json.Unmarshal(data, &m); err != nil {
				continue
			}

			m.Dir = fmt.Sprintf("missions/%s", dirName)
			manifestMap[m.ID] = m
		}
	}

	// Convert map to slice and sort by ID/name to make it stable
	var keys []string
	for k := range manifestMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var manifests []MissionManifest
	for i, k := range keys {
		m := manifestMap[k]
		// Dynamically correct the "MISSION X:" prefix to reflect the 1-based sorted index
		m.Title = correctMissionTitle(m.Title, i+1)
		manifests = append(manifests, m)
	}

	return manifests, nil
}

// correctMissionTitle replaces any existing "MISSION X:" prefix with "MISSION [index]:"
func correctMissionTitle(title string, index int) string {
	prefix := fmt.Sprintf("MISSION %d:", index)
	if missionTitleRegex.MatchString(title) {
		return missionTitleRegex.ReplaceAllString(title, prefix)
	}
	return title
}

// LoadLevelMapData retrieves the raw JSON byte data for a level map file inside a mission.
func (m *MissionManifest) LoadLevelMapData(fileRelativePath string) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/%s", m.Dir, fileRelativePath)

	// Try reading from embedded filesystem. Embedded paths look like "missions/[dirName]/[file]"
	embedPath := fullPath
	if strings.HasPrefix(embedPath, "assets/") {
		embedPath = strings.TrimPrefix(embedPath, "assets/")
	}
	data, err := assets.AssetsFS.ReadFile(embedPath)
	if err == nil {
		return data, nil
	}

	// Fallback: read directly from the OS filesystem path
	return os.ReadFile(fullPath)
}
