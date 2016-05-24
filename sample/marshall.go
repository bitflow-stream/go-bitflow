package sample

import (
	"bufio"
	"fmt"
	"io"
)

const (
	time_col = "time"
	tags_col = "tags"
)

func checkFirstCol(col string) error {
	if col != time_col {
		return fmt.Errorf("Unexpected first column %v, expected %v", col, time_col)
	}
	return nil
}

type Marshaller interface {
	String() string
	WriteHeader(header Header, output io.Writer) error
	WriteSample(sample Sample, header Header, output io.Writer) error
}

type Unmarshaller interface {
	String() string

	// io.EOF is not valid
	ReadHeader(input *bufio.Reader) (Header, error)

	// io.EOF indicates end of stream
	ReadSampleData(header Header, input *bufio.Reader) ([]byte, error)
	ParseSample(header Header, data []byte) (Sample, error)
}

type MetricMarshaller interface {
	String() string
	WriteHeader(header Header, output io.Writer) error
	WriteSample(sample Sample, header Header, output io.Writer) error
	ReadHeader(input *bufio.Reader) (Header, error)
	ReadSampleData(header Header, input *bufio.Reader) ([]byte, error)
	ParseSample(header Header, data []byte) (Sample, error)
}

// Helper type for more concise Write code by avoiding error checks on every
// Write() invokation. Overhead: additional no-op Write()/WriteStr() calls
// after an error has occurred (which is the exception).
type WriteCascade struct {
	Writer io.Writer
	Err    error
}

func (w *WriteCascade) Write(bytes []byte) {
	if w.Err == nil {
		_, w.Err = w.Writer.Write(bytes)
	}
}

func (w *WriteCascade) WriteStr(str string) {
	w.Write([]byte(str))
}

func (w *WriteCascade) WriteByte(b byte) {
	w.Write([]byte{b})
}
