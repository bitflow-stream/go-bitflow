package sample

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

const (
	time_col = "time"
	tags_col = "tags"
)

func checkFirstCol(col string) error {
	if col != time_col {
		if len(col) >= 20 {
			col = col[:20] + "..."
		}
		return fmt.Errorf("First column should be %v, but found: %q", time_col, col)
	}
	return nil
}

func detectFormat(input *bufio.Reader) (Unmarshaller, error) {
	peekNum := len(time_col) + 1
	peeked, err := input.Peek(peekNum)
	if err == bufio.ErrBufferFull {
		err = errors.New("IO buffer is too small to auto-detect input stream format")
	}
	if err != nil {
		return nil, err
	}
	switch peeked[peekNum-1] {
	case csv_separator:
		return new(CsvMarshaller), nil
	case binary_separator:
		return new(BinaryMarshaller), nil
	default:
		return nil, errors.New("Failed to auto-detect format of stream starting with: " + string(peeked))
	}
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
