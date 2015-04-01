package iocontrol

import (
	"sync/atomic"
	"time"

	"github.com/benbjohnson/clock"
)

type rateLimiter struct {
	limitPerSec int
	resolution  time.Duration

	time clock.Clock // YAGNI wrapper for YAGNI deterministic testing

	batchDone int64
	lastBatch time.Time

	// can be modified concurrently
	maxPerBatch int64
}

func newRateLimiter(perSec int, maxBurst time.Duration) *rateLimiter {
	maxPerBatch := int64(perSec / int(time.Second/maxBurst))
	return &rateLimiter{
		limitPerSec: perSec,
		resolution:  maxBurst,
		time:        clock.New(),
		maxPerBatch: maxPerBatch,
	}
}

func (r *rateLimiter) CanDo() (canDo int) {
	perBatch := atomic.LoadInt64(&r.maxPerBatch)
	canDo = int(perBatch - r.batchDone)
	if canDo < 0 {
		return 0
	}
	return canDo
}

func (r *rateLimiter) Did(n int) {
	r.batchDone += int64(n)
}

func (r *rateLimiter) SetRate(perSec int) {
	maxPerBatch := int64(perSec / int(time.Second/r.resolution))
	atomic.StoreInt64(&r.maxPerBatch, maxPerBatch)
}

func (r *rateLimiter) Limit() {
	nextBatch := r.lastBatch.Add(r.resolution)
	durationToNextBatch := nextBatch.Sub(r.time.Now())

	r.time.Sleep(durationToNextBatch)

	r.lastBatch = r.time.Now()
	r.batchDone = 0
}
