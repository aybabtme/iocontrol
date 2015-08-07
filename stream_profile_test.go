package iocontrol

import (
	"io"
	"math"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestProfile(t *testing.T) {

	wantProfile := StreamProfile{
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

	res := make(chan StreamProfile, 1)
	go func() {
		w, r, done := profile(clk, sleepWrite, sleepRead)
		io.Copy(w, r)
		res <- done()
	}()

	var gotProfile StreamProfile
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

type readFunc func([]byte) (int, error)

func (r readFunc) Read(p []byte) (int, error) { return r(p) }

type writeFunc func([]byte) (int, error)

func (w writeFunc) Write(p []byte) (int, error) { return w(p) }
