package ids

import (
	"crypto/rand"
	"crypto/sha256"
	"strings"
	"time"

	"github.com/google/uuid"
)

func New() uuid.UUID {
	return newAt(time.Now())
}

func NewPrefixed(prefix string) string {
	return newPrefixedAt(prefix, time.Now())
}

func newPrefixedAt(prefix string, now time.Time) string {
	id := newAt(now)
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "_")
	if prefix == "" {
		return id.String()
	}
	return prefix + "_" + id.String()
}

func newAt(now time.Time) uuid.UUID {
	var id uuid.UUID
	milliseconds := uint64(now.UnixMilli())
	id[0] = byte(milliseconds >> 40)
	id[1] = byte(milliseconds >> 32)
	id[2] = byte(milliseconds >> 24)
	id[3] = byte(milliseconds >> 16)
	id[4] = byte(milliseconds >> 8)
	id[5] = byte(milliseconds)

	first := sha256.Sum256([]byte(rand.Text()))
	second := sha256.Sum256([]byte(rand.Text()))
	randomBytes := append(first[:], second[:]...)
	id[6] = 0x70 | randomBytes[0]&0x0f
	id[7] = randomBytes[1]
	id[8] = 0x80 | randomBytes[2]&0x3f
	copy(id[9:], randomBytes[3:10])
	return id
}
