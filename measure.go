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

// MeasuredReaderAt wraps an io.ReaderAt and tracks how many bytes are read from it.
type MeasuredReaderAt struct {
	wrap io.ReaderAt
	rate *rateCounter
}

// NewMeasuredReaderAt wraps a ReaderAt.
func NewMeasuredReaderAt(r io.ReaderAt) *MeasuredReaderAt {
	return &MeasuredReaderAt{wrap: r, rate: newCounter()}
}

// BytesPer tells the rate per period at which bytes were read since last measurement.
func (m *MeasuredReaderAt) BytesPer(perPeriod time.Duration) uint64 {
	return uint64(m.rate.Rate(perPeriod))
}

// BytesPerSec tells the rate per second at which bytes were read since last measurement.
func (m *MeasuredReaderAt) BytesPerSec() uint64 {
	return uint64(m.rate.Rate(time.Second))
}

// Total number of bytes that have been read.
func (m *MeasuredReaderAt) Total() int {
	return m.rate.Total()
}

func (m *MeasuredReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = m.wrap.ReadAt(p, off)
	m.rate.Add(n)
	return n, err
}

// MeasuredWriterAt wraps an io.WriterAt and tracks how many bytes are written to it.
type MeasuredWriterAt struct {
	wrap io.WriterAt
	rate *rateCounter
}

// NewMeasuredWriterAt wraps a WriterAt.
func NewMeasuredWriterAt(w io.WriterAt) *MeasuredWriterAt {
	return &MeasuredWriterAt{wrap: w, rate: newCounter()}
}

// BytesPer tells the rate per period at which bytes were written since last measurement.
func (m *MeasuredWriterAt) BytesPer(perPeriod time.Duration) uint64 {
	return uint64(m.rate.Rate(perPeriod))
}

// BytesPerSec tells the rate per second at which bytes were written since last measurement.
func (m *MeasuredWriterAt) BytesPerSec() uint64 {
	return uint64(m.rate.Rate(time.Second))
}

// Total number of bytes that have been written.
func (m *MeasuredWriterAt) Total() int {
	return m.rate.Total()
}

func (m *MeasuredWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = m.wrap.WriteAt(p, off)
	m.rate.Add(n)
	return n, err
}
