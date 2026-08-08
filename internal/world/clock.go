package world

import (
	"fmt"

	"github.com/vikash-paf/derelict-facility/internal/core"
)

// Season represents the current seasonal configuration controlling daylight hours.
type Season uint8

const (
	SeasonSpring Season = iota
	SeasonSummer
	SeasonAutumn
	SeasonWinter
)

func (s Season) String() string {
	switch s {
	case SeasonSpring:
		return "SPRING"
	case SeasonSummer:
		return "SUMMER"
	case SeasonAutumn:
		return "AUTUMN"
	case SeasonWinter:
		return "WINTER"
	default:
		return "SPRING"
	}
}

// SeasonConfig defines sunrise, sunset, and daylight thresholds for a season.
type SeasonConfig struct {
	SunriseHour int // Hour when dawn starts (e.g. 6 = 06:00)
	SunsetHour  int // Hour when dusk starts (e.g. 18 = 18:00)
}

// Default Season configurations (Configurable in code!)
var SeasonConfigs = map[Season]SeasonConfig{
	SeasonSpring: {SunriseHour: 6, SunsetHour: 18}, // Equal Day / Night (12h daylight)
	SeasonSummer: {SunriseHour: 4, SunsetHour: 21}, // Long Days / Short Nights (17h daylight)
	SeasonAutumn: {SunriseHour: 7, SunsetHour: 17}, // Shorter Days / Cool Twilight (10h daylight)
	SeasonWinter: {SunriseHour: 8, SunsetHour: 16}, // Long Nights / Short Days (8h daylight)
}

// DaylightColors define the color palette transitions across time of day.
var (
	NightColor  = core.Color{R: 40, G: 60, B: 110, A: 255}   // Deep lunar blue tint
	DawnColor   = core.Color{R: 255, G: 180, B: 120, A: 255} // Warm golden amber
	MiddayColor = core.Color{R: 255, G: 255, B: 240, A: 255} // Bright crisp white
	DuskColor   = core.Color{R: 220, G: 110, B: 150, A: 255} // Crimson violet dusk
)

// FacilityClock tracks in-game time, day cycle, and seasonal daylight curves.
type FacilityClock struct {
	TotalTicks uint64
	Day        int
	Season     Season
}

func NewFacilityClock() *FacilityClock {
	return &FacilityClock{
		TotalTicks: 360, // Start at 06:00 AM Day 1
		Day:        1,
		Season:     SeasonSpring,
	}
}

// Tick advances time by 1 in-game minute per tick (60 ticks = 1 in-game hour).
func (c *FacilityClock) Tick() {
	c.TotalTicks++
	// Every 1440 minutes (24 hours), advance a day
	if c.TotalTicks%1440 == 0 {
		c.Day++
		// Every 7 days, cycle season
		if c.Day%7 == 1 {
			c.Season = (c.Season + 1) % 4
		}
	}
}

// Hour returns the current hour of the day (0-23).
func (c *FacilityClock) Hour() int {
	return int((c.TotalTicks % 1440) / 60)
}

// Minute returns the current minute of the hour (0-59).
func (c *FacilityClock) Minute() int {
	return int(c.TotalTicks % 60)
}

// FormatTime returns a clean formatted time string (e.g., "08:30 AM | DAY 01 | SPRING")
func (c *FacilityClock) FormatTime() string {
	h := c.Hour()
	m := c.Minute()
	ampm := "AM"
	displayHour := h
	if displayHour >= 12 {
		ampm = "PM"
		if displayHour > 12 {
			displayHour -= 12
		}
	}
	if displayHour == 0 {
		displayHour = 12
	}
	return fmt.Sprintf("%02d:%02d %s | DAY %02d | %s", displayHour, m, ampm, c.Day, c.Season)
}

// GetSunlightColor returns the current ambient sunlight color based on season and time.
func (c *FacilityClock) GetSunlightColor() core.Color {
	config := SeasonConfigs[c.Season]
	hour := float64(c.Hour()) + float64(c.Minute())/60.0

	sunrise := float64(config.SunriseHour)
	sunset := float64(config.SunsetHour)
	midday := (sunrise + sunset) / 2.0

	// Night (before sunrise - 1h or after sunset + 1h)
	if hour < (sunrise-1.0) || hour > (sunset+1.0) {
		return NightColor
	}

	// Dawn transition: (sunrise - 1h) to sunrise
	if hour >= (sunrise-1.0) && hour < sunrise {
		t := hour - (sunrise - 1.0)
		return core.LerpColor(NightColor, DawnColor, t)
	}

	// Morning transition: sunrise to midday
	if hour >= sunrise && hour < midday {
		t := (hour - sunrise) / (midday - sunrise)
		return core.LerpColor(DawnColor, MiddayColor, t)
	}

	// Afternoon transition: midday to sunset
	if hour >= midday && hour < sunset {
		t := (hour - midday) / (sunset - midday)
		return core.LerpColor(MiddayColor, DuskColor, t)
	}

	// Dusk transition: sunset to (sunset + 1h)
	if hour >= sunset && hour <= (sunset+1.0) {
		t := hour - sunset
		return core.LerpColor(DuskColor, NightColor, t)
	}

	return NightColor
}
