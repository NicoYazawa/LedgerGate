package clock

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type systemClock struct{}

type Fake struct {
	mu  sync.RWMutex
	now time.Time
}

func System() Clock {
	return systemClock{}
}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

func NewFake(now time.Time) *Fake {
	return &Fake{now: now}
}

func (fake *Fake) Now() time.Time {
	fake.mu.RLock()
	defer fake.mu.RUnlock()
	return fake.now
}

func (fake *Fake) Advance(duration time.Duration) {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	fake.now = fake.now.Add(duration)
}

func (fake *Fake) Set(now time.Time) {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	fake.now = now
}
