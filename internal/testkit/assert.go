//go:build integration

package testkit

import (
	"testing"
	"time"
)

func Eventually(t testing.TB, timeout time.Duration, condition func() bool, message string) {
	t.Helper()
	if condition() {
		return
	}
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timer.C:
			t.Fatal(message)
		}
	}
}
