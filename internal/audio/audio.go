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
	SoundHeartbeat       SoundID = "heartbeat"
)

type AudioManager struct {
	initialized bool
	muted       bool
	sounds      map[SoundID]rl.Sound
	music       rl.Music
	musicLoaded bool
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
		SoundHeartbeat:       "audio/heartbeat.wav",
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

	// Try loading ambient music from assets if present
	musicData, err := assets.AssetsFS.ReadFile("audio/ambient.wav")
	if err == nil && len(musicData) > 0 {
		a.music = rl.LoadMusicStreamFromMemory(".wav", musicData, int32(len(musicData)))
		if a.music.CtxData != nil {
			a.musicLoaded = true
			rl.PlayMusicStream(a.music)
			rl.SetMusicVolume(a.music, 0.5)
		}
	}
}

func (a *AudioManager) Update() {
	if !a.initialized || a.muted || !a.musicLoaded {
		return
	}
	rl.UpdateMusicStream(a.music)
}

func (a *AudioManager) SetMusicVolume(volume float32) {
	if !a.initialized || !a.musicLoaded {
		return
	}
	rl.SetMusicVolume(a.music, volume)
}

func (a *AudioManager) SetMusicPitch(pitch float32) {
	if !a.initialized || !a.musicLoaded {
		return
	}
	rl.SetMusicPitch(a.music, pitch)
}

func (a *AudioManager) Play(id SoundID) {
	if !a.initialized || a.muted {
		return
	}
	if sound, exists := a.sounds[id]; exists {
		rl.PlaySound(sound)
	}
}

// ToggleMute flips the mute state and returns the new state.
func (a *AudioManager) ToggleMute() bool {
	a.muted = !a.muted
	if a.muted && a.musicLoaded {
		rl.PauseMusicStream(a.music)
	} else if !a.muted && a.musicLoaded {
		rl.ResumeMusicStream(a.music)
	}
	return a.muted
}

// IsMuted returns true when sound is muted.
func (a *AudioManager) IsMuted() bool {
	return a.muted
}

func (a *AudioManager) Close() {
	if !a.initialized {
		return
	}
	if a.musicLoaded {
		rl.UnloadMusicStream(a.music)
		a.musicLoaded = false
	}
	for _, sound := range a.sounds {
		rl.UnloadSound(sound)
	}
	rl.CloseAudioDevice()
	a.initialized = false
}
