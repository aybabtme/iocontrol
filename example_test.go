package iocontrol_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/aybabtme/iocontrol"
	"github.com/dustin/go-humanize"
	"io"
	"io/ioutil"
	"log"
	"time"
)

func Example_ThrottledReader() {

	totalSize := 10 * iocontrol.KiB
	readPerSec := 100 * iocontrol.KiB
	maxBurst := 10 * time.Millisecond

	input := randBytes(totalSize)
	src := bytes.NewReader(input)
	measured := iocontrol.NewMeasuredReader(src)
	throttled := iocontrol.ThrottledReader(measured, readPerSec, maxBurst)

	done := make(chan []byte)
	go func() {
		start := time.Now()
		output, err := ioutil.ReadAll(throttled)
		fmt.Printf("done in %.1fs", time.Since(start).Seconds())
		if err != nil {
			log.Fatalf("error reading: %v", err)
		}
		done <- output
	}()

	for {
		select {
		case <-time.Tick(time.Millisecond * 10):
			log.Printf("reading at %s/s", humanize.IBytes(measured.BytesPerSec()))

		case output := <-done:
			if !bytes.Equal(input, output) {
				log.Print("==== input ====\n", hex.Dump(input))
				log.Print("==== output ====\n", hex.Dump(output))
				log.Fatalf("mismatch between input and output")
			}
			return
		}
	}

	// Output:
	// done in 0.1s
}

func Example_ThrottledWriter() {

	totalSize := 10 * iocontrol.KiB
	readPerSec := 100 * iocontrol.KiB
	maxBurst := 10 * time.Millisecond

	input := randBytes(totalSize)
	src := bytes.NewReader(input)
	dst := bytes.NewBuffer(nil)
	measured := iocontrol.NewMeasuredWriter(dst)
	throttled := iocontrol.ThrottledWriter(measured, readPerSec, maxBurst)

	done := make(chan []byte)
	go func() {
		start := time.Now()
		_, err := io.Copy(throttled, src)
		fmt.Printf("done in %.1fs", time.Since(start).Seconds())
		if err != nil {
			log.Fatalf("error writing: %v", err)
		}
		done <- dst.Bytes()
	}()

	for {
		select {
		case <-time.Tick(time.Millisecond * 10):
			log.Printf("writing at %s/s", humanize.IBytes(measured.BytesPerSec()))

		case output := <-done:
			if !bytes.Equal(input, output) {
				log.Print("==== input ====\n", hex.Dump(input))
				log.Print("==== output ====\n", hex.Dump(output))
				log.Fatalf("mismatch between input and output")
			}
			return
		}
	}

	// Output:
	// done in 0.1s
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}
