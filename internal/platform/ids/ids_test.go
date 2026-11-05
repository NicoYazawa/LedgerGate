package ids

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewReturnsUUIDv7(t *testing.T) {
	id := New()

	if id.Version() != 7 {
		t.Errorf("Version() = %d, want 7", id.Version())
	}
	if id == uuid.Nil {
		t.Error("New() returned nil UUID")
	}
}

func TestUUIDv7TimestampOrdering(t *testing.T) {
	earlier := time.Date(2026, 11, 5, 8, 0, 0, 100_000_000, time.UTC)
	later := earlier.Add(2 * time.Millisecond)

	first := newAt(earlier)
	second := newAt(later)

	if first.Version() != 7 || second.Version() != 7 {
		t.Fatalf("versions = %d and %d, want 7", first.Version(), second.Version())
	}
	firstSeconds, firstNanoseconds := first.Time().UnixTime()
	if got := time.Unix(firstSeconds, firstNanoseconds).UTC(); !got.Equal(earlier) {
		t.Errorf("first time = %s, want %s", got, earlier)
	}
	secondSeconds, secondNanoseconds := second.Time().UnixTime()
	if got := time.Unix(secondSeconds, secondNanoseconds).UTC(); !got.Equal(later) {
		t.Errorf("second time = %s, want %s", got, later)
	}
	if strings.Compare(first.String(), second.String()) >= 0 {
		t.Errorf("IDs are not time ordered: %s >= %s", first, second)
	}
}

func TestPrefixedIDsUseStablePrefix(t *testing.T) {
	fixed := time.Date(2026, 11, 5, 8, 30, 0, 0, time.UTC)

	withPrefix := newPrefixedAt("req_", fixed)
	withoutPrefix := newPrefixedAt("", fixed)

	if !strings.HasPrefix(withPrefix, "req_") {
		t.Errorf("prefixed ID = %q, want req_ prefix", withPrefix)
	}
	if strings.HasPrefix(withPrefix, "req__") {
		t.Errorf("prefixed ID = %q, want one underscore", withPrefix)
	}
	if strings.Contains(withoutPrefix, "_") {
		t.Errorf("unprefixed ID = %q, want UUID only", withoutPrefix)
	}
	if _, err := uuid.Parse(strings.TrimPrefix(withPrefix, "req_")); err != nil {
		t.Errorf("prefixed UUID is invalid: %v", err)
	}
}

func TestNewPrefixedUsesPublicGenerator(t *testing.T) {
	value := NewPrefixed("att")
	parsed, err := uuid.Parse(strings.TrimPrefix(value, "att_"))
	if err != nil {
		t.Fatalf("uuid.Parse() error = %v", err)
	}
	if parsed.Version() != 7 {
		t.Errorf("Version() = %d, want 7", parsed.Version())
	}
}
