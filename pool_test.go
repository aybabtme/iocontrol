package iocontrol

import (
	"bytes"
	"io"
	"io/ioutil"
	"sync"
	"testing"
	"time"
)

// writer

func TestWriterPool(t *testing.T) {

	writePerSec := 10 * KiB
	maxBurst := 5 * time.Millisecond

	var wg sync.WaitGroup
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	pool := NewWriterPool(writePerSec, maxBurst)

	mwGlobal := NewMeasuredWriter(ioutil.Discard)
	mwA := NewMeasuredWriter(mwGlobal)
	mwB := NewMeasuredWriter(mwGlobal)

	wg.Add(1)
	go func() {
		defer wg.Done()
		useWriter(pool.Get(mwA))
	}()

	time.Sleep(time.Millisecond * 10)
	if want, got := 1, pool.Len(); want != got {
		t.Errorf("want Len %d, got %d", want, got)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		useWriter(pool.Get(mwB))
	}()

	time.Sleep(time.Millisecond * 10)
	if want, got := 2, pool.Len(); want != got {
		t.Errorf("want Len %d, got %d", want, got)
	}

	assertWriteRate(t, writePerSec, mwGlobal, 5, 20*time.Millisecond)
	assertWriteRate(t, writePerSec/2, mwA, 5, 20*time.Millisecond)
	assertWriteRate(t, writePerSec/2, mwB, 5, 20*time.Millisecond)

	oldRate := pool.SetRate(1 * GiB)
	if oldRate != writePerSec {
		t.Errorf("want old rate %v, got %v", writePerSec, oldRate)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		panic("too long!")
	}

	if want, got := 0, pool.Len(); want != got {
		t.Errorf("want Len %d, got %d", want, got)
	}

	// check that we can restart the process

	pool.SetRate(writePerSec)

	var reusedG sync.WaitGroup
	reusedDone := make(chan struct{})
	go func() {
		reusedG.Wait()
		close(reusedDone)
	}()
	loneWriter := NewMeasuredWriter(ioutil.Discard)
	reusedG.Add(1)
	go func() {
		defer reusedG.Done()
		useWriter(pool.Get(loneWriter))
	}()

	time.Sleep(time.Millisecond * 10)
	if want, got := 1, pool.Len(); want != got {
		t.Errorf("want Len %d, got %d", want, got)
	}

	assertWriteRate(t, writePerSec, loneWriter, 5, 20*time.Millisecond)

	// make it finish
	pool.SetRate(1 * GiB)

	select {
	case <-reusedDone:
	case <-time.After(time.Second * 2):
		panic("too long!")
	}
}

func useWriter(w io.Writer, release func()) {
	totalSize := 10 * MiB
	defer release()
	src := bytes.NewReader(make([]byte, totalSize))
	io.Copy(w, src)
}

// reader

func TestReaderPool(t *testing.T) {

	totalSize := 10 * MiB
	readPerSec := 10 * KiB
	maxBurst := 5 * time.Millisecond
	src := bytes.NewReader(make([]byte, totalSize))

	var wg sync.WaitGroup
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	pool := NewReaderPool(readPerSec, maxBurst)

	mrGlobal := NewMeasuredReader(src)
	mrA := NewMeasuredReader(mrGlobal)
	mrB := NewMeasuredReader(mrGlobal)

	wg.Add(1)
	go func() {
		defer wg.Done()
		useReader(pool.Get(mrA))
	}()

	time.Sleep(time.Millisecond * 10)
	if want, got := 1, pool.Len(); want != got {
		t.Errorf("want Len %d, got %d", want, got)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		useReader(pool.Get(mrB))
	}()

	time.Sleep(time.Millisecond * 10)
	if want, got := 2, pool.Len(); want != got {
		t.Errorf("want Len %d, got %d", want, got)
	}

	assertReadRate(t, readPerSec, mrGlobal, 5, 20*time.Millisecond)
	assertReadRate(t, readPerSec/2, mrA, 5, 20*time.Millisecond)
	assertReadRate(t, readPerSec/2, mrB, 5, 20*time.Millisecond)

	oldRate := pool.SetRate(1 * GiB)
	if oldRate != readPerSec {
		t.Errorf("want old rate %v, got %v", readPerSec, oldRate)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		panic("too long!")
	}

	if want, got := 0, pool.Len(); want != got {
		t.Errorf("want Len %d, got %d", want, got)
	}

	// check that we can restart the process

	pool.SetRate(readPerSec)

	var reusedG sync.WaitGroup
	reusedDone := make(chan struct{})
	go func() {
		reusedG.Wait()
		close(reusedDone)
	}()
	src.Seek(0, 0)
	loneReader := NewMeasuredReader(src)
	reusedG.Add(1)
	go func() {
		defer reusedG.Done()
		useReader(pool.Get(loneReader))
	}()

	time.Sleep(time.Millisecond * 10)
	if want, got := 1, pool.Len(); want != got {
		t.Errorf("want Len %d, got %d", want, got)
	}

	assertReadRate(t, readPerSec, loneReader, 5, 20*time.Millisecond)

	// make it finish
	pool.SetRate(1 * GiB)

	select {
	case <-reusedDone:
	case <-time.After(time.Second * 2):
		panic("too long!")
	}
}

func useReader(r io.Reader, release func()) {
	defer release()
	io.Copy(ioutil.Discard, r)
}
