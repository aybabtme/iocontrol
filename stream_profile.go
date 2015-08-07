package iocontrol

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/benbjohnson/clock"
)

// StreamProfile contains information about the timings involved in
// the exchange of bytes between an io.Writer and io.Reader.
type StreamProfile struct {
	WaitRead  time.Duration
	WaitWrite time.Duration
	Total     time.Duration
}

// Profile will wrap a writer and reader pair and profile where
// time is spent: writing or reading. The result is returned when
// the `done` func is called. The `done` func can be called multiple
// times.
func Profile(w io.Writer, r io.Reader) (pw io.Writer, pr io.Reader, done func() StreamProfile) {
	return profile(clock.New(), w, r)
}

func profile(clk clock.Clock, w io.Writer, r io.Reader) (io.Writer, io.Reader, func() StreamProfile) {

	preciseWriter := &preciseTimedWriter{clk: clk, w: w}
	preciseReader := &preciseTimedReader{clk: clk, r: r}

	start := clk.Now()
	return preciseWriter, preciseReader, func() StreamProfile {
		return StreamProfile{
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
