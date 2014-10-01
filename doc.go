/*
Package iocontrol offers `io.Writer` and `io.Reader` implementations
that allow one to measure and throttle the rate at which data
is transferred.
*/
package iocontrol

// Orders of magnitude of data, in kibibyte (powers of 2, or multiples of 1024).
// See https://en.wikipedia.org/wiki/Kibibyte.
const (
	KiB = 1 << 10
	MiB = 1 << 20
	GiB = 1 << 30
)
