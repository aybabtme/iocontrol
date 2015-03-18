package iocontrol

import (
	"bytes"
	"io"
	"io/ioutil"
	"math"
	"sort"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
)

func TestSetWriteRate(t *testing.T) {

	totalSize := 10 * MiB
	writePerSec := 10 * KiB
	maxBurst := 5 * time.Millisecond

	mw := NewMeasuredWriter(ioutil.Discard)
	tw := ThrottledWriter(mw, writePerSec, maxBurst)

	setRate := func(n int) {
		tw.SetRate(n)
		// discard a measurement
		time.Sleep(time.Millisecond * 10)
		mw.BytesPerSec()
	}

	src := bytes.NewReader(make([]byte, totalSize))
	done := make(chan struct{})
	go func() {
		io.Copy(tw, src)
		close(done)
	}()

	time.Sleep(time.Millisecond * 10)
	mw.BytesPerSec()

	setRate(10 * MiB)
	assertWriteRate(t, 10*MiB, mw, 5, 20*time.Millisecond)

	setRate(1 * MiB)
	assertWriteRate(t, 1*MiB, mw, 5, 20*time.Millisecond)

	setRate(1 * GiB)

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		panic("too long!")
	}
}

func TestSetReadRate(t *testing.T) {

	totalSize := 10 * MiB
	writePerSec := 10 * KiB
	maxBurst := 5 * time.Millisecond

	src := bytes.NewReader(make([]byte, totalSize))

	mr := NewMeasuredReader(src)
	tr := ThrottledReader(mr, writePerSec, maxBurst)

	setRate := func(n int) {
		tr.SetRate(n)
		// discard a measurement
		time.Sleep(time.Millisecond * 10)
		mr.BytesPerSec()
	}

	done := make(chan struct{})
	go func() {
		io.Copy(ioutil.Discard, tr)
		close(done)
	}()

	time.Sleep(time.Millisecond * 10)
	mr.BytesPerSec()

	setRate(10 * MiB)
	assertReadRate(t, 10*MiB, mr, 5, 20*time.Millisecond)

	setRate(1 * MiB)
	assertReadRate(t, 1*MiB, mr, 5, 20*time.Millisecond)

	setRate(1 * GiB)

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		panic("too long!")
	}
}

func assertWriteRate(t *testing.T, approxPerSec int, mw *MeasuredWriter, measures int, sleep time.Duration) {
	eachSleep := sleep / time.Duration(measures)

	rates := make([]int, measures)
	for i := 0; i < measures; i++ {
		time.Sleep(eachSleep)
		rates[i] = int(mw.BytesPerSec())
	}
	sort.Ints(rates)
	median := float64(rates[len(rates)/2])
	if len(rates)%2 != 0 {
		// average of two middle ones
		median = (median + float64(rates[(len(rates)/2)+1])) / 2
	}

	approx := float64(approxPerSec)

	percentDiff := math.Min(approx, median) / math.Max(approx, median)

	t.Logf("%.1f%% similarity", percentDiff*100)
	t.Logf("want rate ~%s", humanize.Bytes(uint64(approxPerSec)))
	t.Logf(" got rate ~%s", humanize.Bytes(uint64(median)))

	if percentDiff < 0.8 {
		t.Errorf("!! too different!")
	}
}

func assertReadRate(t *testing.T, approxPerSec int, mw *MeasuredReader, measures int, sleep time.Duration) {
	eachSleep := sleep / time.Duration(measures)

	rates := make([]int, measures)
	for i := 0; i < measures; i++ {
		time.Sleep(eachSleep)
		rates[i] = int(mw.BytesPerSec())
	}
	sort.Ints(rates)
	median := float64(rates[len(rates)/2])
	if len(rates)%2 != 0 {
		// average of two middle ones
		median = (median + float64(rates[(len(rates)/2)+1])) / 2
	}

	approx := float64(approxPerSec)

	percentDiff := math.Min(approx, median) / math.Max(approx, median)
	t.Logf("%.1f%% similarity", percentDiff*100)
	t.Logf("want rate ~%s", humanize.Bytes(uint64(approxPerSec)))
	t.Logf(" got rate ~%s", humanize.Bytes(uint64(median)))

	if percentDiff < 0.8 {
		t.Errorf("!! too different!")
	}
}
