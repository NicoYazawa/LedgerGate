package clock

import (
	"testing"
	"time"
)

func TestFakeClockControlsTime(t *testing.T) {
	start := time.Date(2026, 11, 5, 9, 0, 0, 0, time.UTC)
	fake := NewFake(start)
	var clock Clock = fake

	if got := clock.Now(); !got.Equal(start) {
		t.Errorf("Now() = %s, want %s", got, start)
	}
	fake.Advance(90 * time.Second)
	if got := clock.Now(); !got.Equal(start.Add(90 * time.Second)) {
		t.Errorf("Now() after Advance = %s", got)
	}
	replacement := start.Add(-time.Hour)
	fake.Set(replacement)
	if got := clock.Now(); !got.Equal(replacement) {
		t.Errorf("Now() after Set = %s, want %s", got, replacement)
	}
}

func TestSystemClockTracksWallTime(t *testing.T) {
	before := time.Now()
	got := System().Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("System().Now() = %s, outside [%s, %s]", got, before, after)
	}
}
