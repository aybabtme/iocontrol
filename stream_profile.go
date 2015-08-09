package iocontrol

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/benbjohnson/clock"
)

// TimeProfile contains information about the timings involved in
// the exchange of bytes between an io.Writer and io.Reader.
type TimeProfile struct {
	WaitRead  time.Duration
	WaitWrite time.Duration
	Total     time.Duration
}

// Profile will wrap a writer and reader pair and profile where
// time is spent: writing or reading. The result is returned when
// the `done` func is called. The `done` func can be called multiple
// times.
//
// There is a small performance overhead of ~µs per Read/Write call.
// This is negligible in most I/O workloads. If the overhead is too
// much for your needs, use the `ProfileSample` call.
func Profile(w io.Writer, r io.Reader) (pw io.Writer, pr io.Reader, done func() TimeProfile) {
	return profile(clock.New(), w, r)
}

func profile(clk clock.Clock, w io.Writer, r io.Reader) (io.Writer, io.Reader, func() TimeProfile) {

	preciseWriter := &preciseTimedWriter{clk: clk, w: w}
	preciseReader := &preciseTimedReader{clk: clk, r: r}

	start := clk.Now()
	return preciseWriter, preciseReader, func() TimeProfile {
		return TimeProfile{
			Total:     clk.Now().Sub(start),
			WaitRead:  preciseReader.WaitRead(),
			WaitWrite: preciseWriter.WaitWrite(),
		}
	}
}

type preciseTimedReader struct {
	clk   clock.Clock
	r     io.Reader
	sumNS int64
}

func (t *preciseTimedReader) WaitRead() time.Duration {
	return time.Duration(atomic.LoadInt64(&t.sumNS))
}

func (t *preciseTimedReader) Read(p []byte) (int, error) {
	start := t.clk.Now()
	n, err := t.r.Read(p)
	atomic.AddInt64(&t.sumNS, t.clk.Now().Sub(start).Nanoseconds())
	return n, err
}

type preciseTimedWriter struct {
	clk   clock.Clock
	w     io.Writer
	sumNS int64
}

func (t *preciseTimedWriter) WaitWrite() time.Duration {
	return time.Duration(atomic.LoadInt64(&t.sumNS))
}

func (t *preciseTimedWriter) Write(p []byte) (int, error) {
	start := t.clk.Now()
	n, err := t.w.Write(p)
	atomic.AddInt64(&t.sumNS, t.clk.Now().Sub(start).Nanoseconds())
	return n, err
}

// sampling, high performance profiler

const (
	stateRuntime uint32 = iota
	stateBlocked
)

// ProfileSample will wrap a writer and reader pair and collect
// samples of where time is spent: writing or reading. The result
// is an approximation that is returned when the `done` func is
// called. The `done` func can be called *only once*.
//
// This call is not as precise as the `Profile` call, but the
// performance overhead is much reduced.
func ProfileSample(w io.Writer, r io.Reader, res time.Duration) (pw io.Writer, pr io.Reader, done func() SamplingProfile) {
	return profileSample(clock.New(), w, r, res)
}

// SamplingProfile samples when a reader and a writer are blocked, or not.
// If sampled at a high enough resolution, the result should give a good
// approximation of the distribution of time. The results are not as
// precise as the result of `Profile`, but the performance overhead
// is much reduced.
type SamplingProfile struct {
	TimeProfile
	Reading    int
	Writing    int
	NotReading int
	NotWriting int
}

func profileSample(clk clock.Clock, w io.Writer, r io.Reader, res time.Duration) (io.Writer, io.Reader, func() SamplingProfile) {
	samplingWriter := &samplingTimeWriter{w: w}
	samplingReader := &samplingTimeReader{r: r}

	start := clk.Now()
	done := make(chan struct{})

	samples := SamplingProfile{}
	go func() {
		ticker := clk.Ticker(res)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:

				isWriting := atomic.LoadUint32(&samplingWriter.state) == stateBlocked
				isReading := atomic.LoadUint32(&samplingReader.state) == stateBlocked

				if isWriting {
					samples.Writing++
				} else {
					samples.NotWriting++
				}
				if isReading {
					samples.Reading++
				} else {
					samples.NotReading++
				}
			case <-done:
				return
			}
		}
	}()

	return samplingWriter, samplingReader, func() SamplingProfile {
		close(done)
		total := clk.Now().Sub(start)
		samples.TimeProfile = TimeProfile{
			Total:     total,
			WaitRead:  time.Duration(float64(samples.Reading) / float64(samples.Reading+samples.NotReading) * float64(total)),
			WaitWrite: time.Duration(float64(samples.Writing) / float64(samples.Writing+samples.NotWriting) * float64(total)),
		}
		return samples
	}
}

type samplingTimeReader struct {
	state uint32
	r     io.Reader
}

func (s *samplingTimeReader) Read(p []byte) (int, error) {
	atomic.StoreUint32(&s.state, stateBlocked)
	n, err := s.r.Read(p)
	atomic.StoreUint32(&s.state, stateRuntime)
	return n, err
}

type samplingTimeWriter struct {
	state uint32
	w     io.Writer
}

func (s *samplingTimeWriter) Write(p []byte) (int, error) {
	atomic.StoreUint32(&s.state, stateBlocked)
	n, err := s.w.Write(p)
	atomic.StoreUint32(&s.state, stateRuntime)
	return n, err
}
