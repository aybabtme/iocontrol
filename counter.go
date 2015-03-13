package iocontrol

import (
	"sync"
	"time"

	"github.com/benbjohnson/clock"
)

type rateCounter struct {
	time  clock.Clock // YAGNI, maybe, idk, how to test this
	mu    sync.RWMutex
	count int

	lastCount int
	lastCheck time.Time
}

func newCounter() *rateCounter {
	return &rateCounter{
		time: clock.New(),
	}
}

func (c *rateCounter) Add(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count += n
	if c.lastCheck.IsZero() {
		c.lastCheck = time.Now()
	}
}

func (c *rateCounter) Total() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}

func (c *rateCounter) Rate(perPeriod time.Duration) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.time.Now()

	between := now.Sub(c.lastCheck)

	changed := c.count - c.lastCount
	rate := float64(changed*int(perPeriod)) / float64(between)

	c.lastCount = c.count
	c.lastCheck = now
	return rate
}
