//go:build integration

package testkit

import (
	"testing"
	"time"
)

func TestEventuallySucceeds(t *testing.T) {
	attempts := 0
	Eventually(t, time.Second, func() bool {
		attempts++
		return attempts == 3
	}, "condition did not become true")
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}
