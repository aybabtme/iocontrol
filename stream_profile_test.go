package iocontrol

import (
	"io"
	"io/ioutil"
	"math"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestProfile(t *testing.T) {

	wantProfile := TimeProfile{
		Total:     2 * time.Second,
		WaitRead:  500 * time.Millisecond,
		WaitWrite: 1500 * time.Millisecond,
	}
	clk := clock.NewMock()

	start := clk.Now()

	sleepRead := readFunc(func(p []byte) (int, error) {
		var err error
		if clk.Now().Sub(start) >= wantProfile.Total {
			err = io.EOF
		}
		clk.Sleep(wantProfile.WaitRead / 1000)
		return len(p), err
	})

	sleepWrite := writeFunc(func(p []byte) (int, error) {
		clk.Sleep(wantProfile.WaitWrite / 1000)
		return len(p), nil
	})

	res := make(chan TimeProfile, 1)
	go func() {
		w, r, done := profile(clk, sleepWrite, sleepRead)
		io.Copy(w, r)
		res <- done()
	}()

	var gotProfile TimeProfile
loop:
	for {
		select {
		case gotProfile = <-res:
			break loop
		default:
			clk.Add(1 * time.Millisecond)
		}
	}

	wantReadRatio := float64(wantProfile.WaitRead) / float64(wantProfile.WaitRead+wantProfile.WaitWrite)
	gotReadRatio := float64(gotProfile.WaitRead) / float64(gotProfile.WaitRead+gotProfile.WaitWrite)

	diff := (math.Max(wantReadRatio, gotReadRatio) - math.Min(wantReadRatio, gotReadRatio)) / math.Max(wantReadRatio, gotReadRatio)
	if diff > 0.05 {
		t.Logf("want=%#v", wantProfile)
		t.Logf(" got=%#v", gotProfile)
		t.Fatalf("profiles are too different: %.2f%% different", 100*diff)
	}
}

func TestProfileSample(t *testing.T) {

	wantProfile := TimeProfile{
		Total:     10 * time.Second,
		WaitRead:  1 * time.Second,
		WaitWrite: 9 * time.Second,
	}
	clk := clock.NewMock()

	start := clk.Now()

	sleepRead := readFunc(func(p []byte) (int, error) {
		var err error
		if clk.Now().Sub(start) >= wantProfile.Total {
			err = io.EOF
		}
		clk.Sleep(wantProfile.WaitRead / 1000)
		return len(p), err
	})

	sleepWrite := writeFunc(func(p []byte) (int, error) {
		clk.Sleep(wantProfile.WaitWrite / 1000)
		return len(p), nil
	})

	res := make(chan TimeProfile, 1)
	go func() {
		w, r, done := profileSample(clk, sleepWrite, sleepRead, time.Millisecond)
		io.Copy(w, r)
		res <- done().TimeProfile
	}()

	var gotProfile TimeProfile
loop:
	for {
		select {
		case gotProfile = <-res:
			break loop
		default:
			clk.Add(1 * time.Millisecond)
		}
	}

	wantReadRatio := float64(wantProfile.WaitRead) / float64(wantProfile.WaitRead+wantProfile.WaitWrite)
	gotReadRatio := float64(gotProfile.WaitRead) / float64(gotProfile.WaitRead+gotProfile.WaitWrite)

	diff := (math.Max(wantReadRatio, gotReadRatio) - math.Min(wantReadRatio, gotReadRatio)) / math.Max(wantReadRatio, gotReadRatio)
	if diff > 0.05 {
		t.Logf("want=%#v", wantProfile)
		t.Logf(" got=%#v", gotProfile)
		t.Fatalf("profiles are too different: %.2f%% different", 100*diff)
	}
}

func BenchmarkNoProfile(b *testing.B) {
	byteCount := int64(1 << 30)
	b.SetBytes(byteCount)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := io.LimitReader(readFunc(func(p []byte) (int, error) {
			return len(p), nil
		}), byteCount)
		writer := ioutil.Discard

		b.StartTimer()
		io.Copy(writer, reader)
		b.StopTimer()
	}
}

func BenchmarkProfile(b *testing.B) {
	byteCount := int64(1 << 30)
	b.SetBytes(byteCount)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := io.LimitReader(readFunc(func(p []byte) (int, error) {
			return len(p), nil
		}), byteCount)
		writer := ioutil.Discard
		pwriter, preader, done := Profile(writer, reader)

		b.StartTimer()
		io.Copy(pwriter, preader)
		b.StopTimer()

		done()
	}
}

func BenchmarkProfileSample(b *testing.B) {
	byteCount := int64(1 << 30)
	b.SetBytes(byteCount)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := io.LimitReader(readFunc(func(p []byte) (int, error) {
			return len(p), nil
		}), byteCount)
		writer := ioutil.Discard

		pwriter, preader, done := ProfileSample(writer, reader, time.Millisecond)

		b.StartTimer()
		io.Copy(pwriter, preader)
		b.StopTimer()

		done()
	}
}

type readFunc func([]byte) (int, error)

func (r readFunc) Read(p []byte) (int, error) { return r(p) }

type writeFunc func([]byte) (int, error)

func (w writeFunc) Write(p []byte) (int, error) { return w(p) }
