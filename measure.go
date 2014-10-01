package iocontrol

import (
	"io"
	"time"
)

// MeasuredWriter wraps a writer and tracks how many bytes are written to it.
type MeasuredWriter struct {
	wrap io.Writer
	rate *rateCounter
}

// NewMeasuredWriter wraps a writer.
func NewMeasuredWriter(w io.Writer) *MeasuredWriter {
	return &MeasuredWriter{wrap: w, rate: newCounter()}
}

// BytesPer tells the rate per period at which bytes were written since last
// measurement.
func (m *MeasuredWriter) BytesPer(perPeriod time.Duration) uint64 {
	return uint64(m.rate.Rate(perPeriod))
}

// BytesPerSec tells the rate per second at which bytes were written since last
// measurement.
func (m *MeasuredWriter) BytesPerSec() uint64 {
	return uint64(m.rate.Rate(time.Second))
}

// Total number of bytes that have been written.
func (m *MeasuredWriter) Total() int {
	return m.rate.Total()
}

func (m *MeasuredWriter) Write(b []byte) (n int, err error) {
	n, err = m.wrap.Write(b)
	m.rate.Add(n)
	return n, err
}

// MeasuredReader wraps a reader and tracks how many bytes are read to it.
type MeasuredReader struct {
	wrap io.Reader
	rate *rateCounter
}

// NewMeasuredReader wraps a reader.
func NewMeasuredReader(r io.Reader) *MeasuredReader {
	return &MeasuredReader{wrap: r, rate: newCounter()}
}

// BytesPer tells the rate per period at which bytes were read since last
// measurement.
func (m *MeasuredReader) BytesPer(perPeriod time.Duration) uint64 {
	return uint64(m.rate.Rate(perPeriod))
}

// BytesPerSec tells the rate per second at which bytes were read since last
// measurement.
func (m *MeasuredReader) BytesPerSec() uint64 {
	return uint64(m.rate.Rate(time.Second))
}

// Total number of bytes that have been read.
func (m *MeasuredReader) Total() int {
	return m.rate.Total()
}

func (m *MeasuredReader) Read(b []byte) (n int, err error) {
	n, err = m.wrap.Read(b)
	m.rate.Add(n)
	return n, err
}
