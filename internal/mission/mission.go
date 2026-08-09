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
	Dir        string      `json:"-"`
}

// MissionLoader loads story campaign manifests from embedded or external files.
type MissionLoader struct{}

func NewMissionLoader() *MissionLoader {
	return &MissionLoader{}
}

// DiscoverMissions scans embedded assets/missions, local OS assets/missions, and local OS `./missions` directory, merging and indexing them correctly.
func (l *MissionLoader) DiscoverMissions() ([]MissionManifest, error) {
	manifests := make(map[string]MissionManifest)

	l.discoverEmbedded(manifests)
	l.discoverLocal("assets/missions", manifests)
	l.discoverLocal("missions", manifests)

	return l.sortAndIndexMissions(manifests), nil
}

func (l *MissionLoader) discoverEmbedded(manifests map[string]MissionManifest) {
	entries, err := assets.AssetsFS.ReadDir("missions")
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		manifest, err := loadManifest("missions", dirName, assets.AssetsFS.ReadFile)
		if err == nil {
			manifests[manifest.ID] = manifest
		}
	}
}

func (l *MissionLoader) discoverLocal(dirPath string, manifests map[string]MissionManifest) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		manifest, err := loadManifest(dirPath, dirName, os.ReadFile)
		if err == nil {
			manifests[manifest.ID] = manifest
		}
	}
}

func loadManifest(dirPath string, dirName string, readFunc func(string) ([]byte, error)) (MissionManifest, error) {
	manifestPath := fmt.Sprintf("%s/%s/mission.json", dirPath, dirName)
	data, err := readFunc(manifestPath)
	if err != nil {
		return MissionManifest{}, err
	}
	var manifest MissionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return MissionManifest{}, err
	}
	manifest.Dir = fmt.Sprintf("%s/%s", dirPath, dirName)
	return manifest, nil
}

func (l *MissionLoader) sortAndIndexMissions(manifests map[string]MissionManifest) []MissionManifest {
	var ids []string
	for id := range manifests {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var ordered []MissionManifest
	for i, id := range ids {
		manifest := manifests[id]
		manifest.Title = correctMissionTitle(manifest.Title, i+1)
		ordered = append(ordered, manifest)
	}
	return ordered
}

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

	embedPath := fullPath
	embedPath, _ = strings.CutPrefix(embedPath, "assets/")

	data, err := assets.AssetsFS.ReadFile(embedPath)
	if err == nil {
		return data, nil
	}

	return os.ReadFile(fullPath)
}
