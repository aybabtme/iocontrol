package iocontrol

import (
	"github.com/benbjohnson/clock"
	"time"
)

type rateLimiter struct {
	limitPerSec int
	resolution  time.Duration

	time clock.Clock // YAGNI wrapper for YAGNI deterministic testing

	maxPerBatch int
	batchDone   int
	lastBatch   time.Time
}

func newRateLimiter(perSec int, maxBurst time.Duration) *rateLimiter {
	maxPerBatch := perSec / int(time.Second/maxBurst)
	return &rateLimiter{
		limitPerSec: perSec,
		resolution:  maxBurst,
		time:        clock.New(),
		maxPerBatch: maxPerBatch,
	}
}

func (r *rateLimiter) CanDo() (canDo int) {
	return r.maxPerBatch - r.batchDone
}

func (r *rateLimiter) Did(n int) {
	r.batchDone += n
}

func (r *rateLimiter) Limit() {
	nextBatch := r.lastBatch.Add(r.resolution)
	durationToNextBatch := nextBatch.Sub(r.time.Now())

	r.time.Sleep(durationToNextBatch)

	r.lastBatch = r.time.Now()
	r.batchDone = 0
}
