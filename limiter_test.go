package iocontrol

import (
	"testing"
	"time"
)

func TestLimiterCanDo(t *testing.T) {
	limiter := newRateLimiter(3*MiB, time.Second)
	limiter.Did(2 * MiB)     // simulate writing
	limiter.SetRate(1 * MiB) // limiter is now less than we've written so far
	canDo := limiter.CanDo()
	if canDo != 0 {
		t.Fatalf("wanted to be able to write nothing, got: %d", canDo)
	}
}
