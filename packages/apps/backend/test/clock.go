package test

import (
	"sync"
	"time"

	rez "github.com/rezible/rezible"
)

// Clock is a controllable rez.Clock for tests. It is safe for concurrent use.
type Clock struct {
	mu  sync.Mutex
	now time.Time
}

var _ rez.Clock = (*Clock)(nil)

func NewClock(start time.Time) *Clock {
	return &Clock{now: start}
}

func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *Clock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}

func (c *Clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}
