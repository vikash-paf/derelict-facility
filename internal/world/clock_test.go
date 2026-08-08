package world

import (
	"testing"
)

func TestFacilityClockSeasons(t *testing.T) {
	clock := NewFacilityClock()

	// Initial time: 06:00 AM, Day 1, Spring
	if clock.Hour() != 6 || clock.Minute() != 0 {
		t.Errorf("expected 06:00, got %02d:%02d", clock.Hour(), clock.Minute())
	}
	if clock.Season != SeasonSpring {
		t.Errorf("expected Spring, got %v", clock.Season)
	}

	// Midday in Spring (12:00) should be MiddayColor
	clock.TotalTicks = 12 * 60
	middayColor := clock.GetSunlightColor()
	if middayColor != MiddayColor {
		t.Errorf("expected MiddayColor at 12:00 Spring, got %+v", middayColor)
	}

	// Midnight in Spring (00:00) should be NightColor
	clock.TotalTicks = 0
	nightColor := clock.GetSunlightColor()
	if nightColor != NightColor {
		t.Errorf("expected NightColor at 00:00 Spring, got %+v", nightColor)
	}

	// Test Winter (Short day: sunrise 8, sunset 16)
	clock.Season = SeasonWinter
	clock.TotalTicks = 7 * 60 // 07:00 AM in Winter is still Night
	winter7am := clock.GetSunlightColor()
	if winter7am != NightColor {
		t.Errorf("expected NightColor at 07:00 AM in Winter, got %+v", winter7am)
	}
}
