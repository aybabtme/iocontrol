package iocontrol

import (
	"io"
	"sync"
	"time"
)

// WriterPool creates instances of iocontrol.ThrottlerWriter that are
// managed such that they collectively do not exceed a certain rate.
//
// The default value of WriterPool is not to be used, create instances
// with `NewWriterPool`.
type WriterPool struct {
	mu       sync.Mutex
	maxRate  int
	maxBurst time.Duration

	givenOut map[ThrottlerWriter]struct{}
}

// NewWriterPool creates a pool that ensures the writers it wraps will
// respect an overall maxRate, with maxBurst resolution. The semantics
// of the wrapped writers are the same as those of using a plain
// ThrottledWriter.
func NewWriterPool(maxRate int, maxBurst time.Duration) *WriterPool {
	return &WriterPool{
		maxRate:  maxRate,
		maxBurst: maxBurst,
		givenOut: make(map[ThrottlerWriter]struct{}),
	}
}

// Get a throttled writer that wraps w.
func (pool *WriterPool) Get(w io.Writer) (writer io.Writer, release func()) {
	// don't export a ThrottlerWriter to prevent users changing the rate
	// and expecting their change to be respected, since we might modify
	// the rate under their feet

	// make the initial rate be 0, the actual rate is
	// set in the call to `setSharedRates`.
	wr := ThrottledWriter(w, 0, pool.maxBurst)

	pool.mu.Lock()
	pool.givenOut[wr] = struct{}{}
	pool.setSharedRates()
	pool.mu.Unlock()

	return wr, func() {
		pool.mu.Lock()
		delete(pool.givenOut, wr)
		pool.setSharedRates()
		pool.mu.Unlock()
	}
}

// SetRate of the pool, updating each given out writer to respect the
// newly set rate. Returns the old rate.
func (pool *WriterPool) SetRate(rate int) int {
	pool.mu.Lock()
	old := pool.maxRate
	pool.maxRate = rate
	pool.setSharedRates()
	pool.mu.Unlock()
	return old
}

// Len is the number of currently given out throttled writers.
func (pool *WriterPool) Len() int {
	pool.mu.Lock()
	l := len(pool.givenOut)
	pool.mu.Unlock()
	return l
}

// must be called with a lock held on `pool.mu`
func (pool *WriterPool) setSharedRates() {
	if len(pool.givenOut) == 0 {
		return
	}
	perSecPerWriter := pool.maxRate / len(pool.givenOut)
	for writer := range pool.givenOut {
		writer.SetRate(perSecPerWriter)
	}
}

// ReaderPool creates instances of iocontrol.ThrottlerReader that are
// managed such that they collectively do not exceed a certain rate.
//
// The default value of ReaderPool is not to be used, create instances
// with `NewReaderPool`.
type ReaderPool struct {
	mu       sync.Mutex
	maxRate  int
	maxBurst time.Duration

	givenOut map[ThrottlerReader]struct{}
}

// NewReaderPool creates a pool that ensures the writers it wraps will
// respect an overall maxRate, with maxBurst resolution. The semantics
// of the wrapped writers are the same as those of using a plain
// ThrottledReader.
func NewReaderPool(maxRate int, maxBurst time.Duration) *ReaderPool {
	return &ReaderPool{
		maxRate:  maxRate,
		maxBurst: maxBurst,
		givenOut: make(map[ThrottlerReader]struct{}),
	}
}

// Get a throttled reader that wraps r.
func (pool *ReaderPool) Get(r io.Reader) (reader io.Reader, release func()) {
	// don't export a ThrottlerReader to prevent users changing the rate
	// and expecting their change to be respected, since we might modify
	// the rate under their feet

	// make the initial rate be 0, the actual rate is
	// set in the call to `setSharedRates`.
	rd := ThrottledReader(r, 0, pool.maxBurst)

	pool.mu.Lock()
	pool.givenOut[rd] = struct{}{}
	pool.setSharedRates()
	pool.mu.Unlock()

	return rd, func() {
		pool.mu.Lock()
		delete(pool.givenOut, rd)
		pool.setSharedRates()
		pool.mu.Unlock()
	}
}

// SetRate of the pool, updating each given out reader to respect the
// newly set rate. Returns the old rate.
func (pool *ReaderPool) SetRate(rate int) int {
	pool.mu.Lock()
	old := pool.maxRate
	pool.maxRate = rate
	pool.setSharedRates()
	pool.mu.Unlock()
	return old
}

// Len is the number of currently given out throttled readers.
func (pool *ReaderPool) Len() int {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return len(pool.givenOut)
}

// must be called with a lock held on `pool.mu`
func (pool *ReaderPool) setSharedRates() {
	if len(pool.givenOut) == 0 {
		return
	}
	perSecPerReader := pool.maxRate / len(pool.givenOut)
	for reader := range pool.givenOut {
		reader.SetRate(perSecPerReader)
	}
}
