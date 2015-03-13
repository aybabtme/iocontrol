package iocontrol

import (
	"io"
	"time"
)

type (
	Throttler interface {
		// SetRate changes the rate at which the throttler allows reads or writes.
		SetRate(perSec int)
	}

	ThrottlerReader interface {
		io.Reader
		Throttler
	}

	ThrottlerWriter interface {
		io.Writer
		Throttler
	}
)

// ThrottledReader ensures that reads to `r` never exceeds a specified rate of
// bytes per second. The `maxBurst` duration changes how often the verification is
// done. The smaller the value, the less bursty, but also the more overhead there
// is to the throttling.
func ThrottledReader(r io.Reader, bytesPerSec int, maxBurst time.Duration) ThrottlerReader {
	return &throttledReader{
		wrap:    r,
		limiter: newRateLimiter(bytesPerSec, maxBurst),
	}
}

type throttledReader struct {
	wrap    io.Reader
	limiter *rateLimiter
}

func (t *throttledReader) Read(b []byte) (n int, err error) {
	canRead := t.limiter.CanDo()
	if len(b) <= canRead {
		// no throttling needed
		n, err = t.wrap.Read(b)
		t.limiter.Did(n)
		return n, err
	}

	if canRead > 0 {
		// read what can be read for this batch
		n, err = t.wrap.Read(b[:canRead])
	}

	t.limiter.Limit()

	// return bytes read and let caller try another read
	return n, err
}

// SetRate changes the rate at which the throttled reader allows reads.
func (t *throttledReader) SetRate(perSec int) {
	t.limiter.SetRate(perSec)
}

// ThrottledWriter ensures that writes to `w` never exceeds a specified rate of
// bytes per second. The `maxBurst` duration changes how often the verification is
// done. The smaller the value, the less bursty, but also the more overhead there
// is to the throttling.
func ThrottledWriter(w io.Writer, bytesPerSec int, maxBurst time.Duration) ThrottlerWriter {
	return &throttledWriter{
		wrap:    w,
		limiter: newRateLimiter(bytesPerSec, maxBurst),
	}
}

type throttledWriter struct {
	wrap    io.Writer
	limiter *rateLimiter
}

func (t *throttledWriter) Write(b []byte) (n int, err error) {
	var m int
	for {
		canWrite := t.limiter.CanDo()
		if len(b[n:]) <= canWrite {
			// no throttling needed
			m, err = t.wrap.Write(b[n:])
			n += m
			t.limiter.Did(m)
			return
		}

		// write what can be writen for this batch
		m, err = t.wrap.Write(b[n : n+canWrite])
		n += m
		if err != nil {
			return
		}
		t.limiter.Limit()
	}
}

// SetRate changes the rate at which the throttled writer allows writes.
func (t *throttledWriter) SetRate(perSec int) {
	t.limiter.SetRate(perSec)
}
