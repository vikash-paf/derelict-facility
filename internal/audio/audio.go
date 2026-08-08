package audio

import (
	"path"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/vikash-paf/derelict-facility/assets"
)

type SoundID string

const (
	SoundFootstep        SoundID = "footstep"
	SoundGeneratorToggle SoundID = "generator_toggle"
	SoundDoorOpen        SoundID = "door_open"
	SoundDoorClose       SoundID = "door_close"
	SoundAccessDenied    SoundID = "access_denied"
	SoundTerminalAccess  SoundID = "terminal_access"
)

type AudioManager struct {
	initialized bool
	sounds      map[SoundID]rl.Sound
}

func NewAudioManager() *AudioManager {
	return &AudioManager{
		sounds: make(map[SoundID]rl.Sound),
	}
}

func (a *AudioManager) Init() {
	if a.initialized {
		return
	}
	rl.InitAudioDevice()
	a.initialized = true

	// Load embedded sound files into Raylib sound buffers
	soundFiles := map[SoundID]string{
		SoundFootstep:        "audio/footstep.wav",
		SoundGeneratorToggle: "audio/generator_toggle.wav",
		SoundDoorOpen:        "audio/door_open.wav",
		SoundDoorClose:       "audio/door_close.wav",
		SoundAccessDenied:    "audio/access_denied.wav",
		SoundTerminalAccess:  "audio/terminal_access.wav",
	}

	for id, filePath := range soundFiles {
		data, err := assets.AssetsFS.ReadFile(filePath)
		if err != nil || len(data) == 0 {
			continue
		}
		ext := path.Ext(filePath)
		wave := rl.LoadWaveFromMemory(ext, data, int32(len(data)))
		if wave.FrameCount > 0 {
			sound := rl.LoadSoundFromWave(wave)
			rl.UnloadWave(wave)
			a.sounds[id] = sound
		}
	}
}

func (a *AudioManager) Play(id SoundID) {
	if !a.initialized {
		return
	}
	if sound, exists := a.sounds[id]; exists {
		rl.PlaySound(sound)
	}
}

func (a *AudioManager) Close() {
	if !a.initialized {
		return
	}
	for _, sound := range a.sounds {
		rl.UnloadSound(sound)
	}
	rl.CloseAudioDevice()
	a.initialized = false
}
