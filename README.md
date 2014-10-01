# iocontrol
--
    import "github.com/aybabtme/iocontrol"

Package iocontrol offers `io.Writer` and `io.Reader` implementations that allow
one to measure and throttle the rate at which data is transferred.

## Usage

```go
const (
	KiB = 1 << 10
	MiB = 1 << 20
	GiB = 1 << 30
)
```
Orders of magnitude of data, in kibibyte (powers of 2, or multiples of 1024).
See https://en.wikipedia.org/wiki/Kibibyte.

#### func  ThrottledReader

```go
func ThrottledReader(r io.Reader, bytesPerSec int, maxBurst time.Duration) io.Reader
```
ThrottledReader ensures that reads to `r` never exceeds a specified rate of
bytes per second. The `maxBurst` duration changes how often the verification is
done. The smaller the value, the less bursty, but also the more overhead there
is to the throttling.

#### func  ThrottledWriter

```go
func ThrottledWriter(w io.Writer, bytesPerSec int, maxBurst time.Duration) io.Writer
```
ThrottledWriter ensures that writes to `w` never exceeds a specified rate of
bytes per second. The `maxBurst` duration changes how often the verification is
done. The smaller the value, the less bursty, but also the more overhead there
is to the throttling.

#### type MeasuredReader

```go
type MeasuredReader struct {
}
```

MeasuredReader wraps a reader and tracks how many bytes are read to it.

#### func  NewMeasuredReader

```go
func NewMeasuredReader(r io.Reader) *MeasuredReader
```
NewMeasuredReader wraps a reader.

#### func (*MeasuredReader) BytesPer

```go
func (m *MeasuredReader) BytesPer(perPeriod time.Duration) uint64
```
BytesPer tells the rate per period at which bytes were read since last
measurement.

#### func (*MeasuredReader) BytesPerSec

```go
func (m *MeasuredReader) BytesPerSec() uint64
```
BytesPerSec tells the rate per second at which bytes were read since last
measurement.

#### func (*MeasuredReader) Read

```go
func (m *MeasuredReader) Read(b []byte) (n int, err error)
```

#### func (*MeasuredReader) Total

```go
func (m *MeasuredReader) Total() int
```
Total number of bytes that have been read.

#### type MeasuredWriter

```go
type MeasuredWriter struct {
}
```

MeasuredWriter wraps a writer and tracks how many bytes are written to it.

#### func  NewMeasuredWriter

```go
func NewMeasuredWriter(w io.Writer) *MeasuredWriter
```
NewMeasuredWriter wraps a writer.

#### func (*MeasuredWriter) BytesPer

```go
func (m *MeasuredWriter) BytesPer(perPeriod time.Duration) uint64
```
BytesPer tells the rate per period at which bytes were written since last
measurement.

#### func (*MeasuredWriter) BytesPerSec

```go
func (m *MeasuredWriter) BytesPerSec() uint64
```
BytesPerSec tells the rate per second at which bytes were written since last
measurement.

#### func (*MeasuredWriter) Total

```go
func (m *MeasuredWriter) Total() int
```
Total number of bytes that have been written.

#### func (*MeasuredWriter) Write

```go
func (m *MeasuredWriter) Write(b []byte) (n int, err error)
```
