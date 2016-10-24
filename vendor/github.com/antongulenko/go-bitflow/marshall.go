package bitflow

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

// Marshaller is an interface for converting Samples and Headers into byte streams.
// The byte streams can be anything including files, network connections, console output,
// or in-memory byte buffers.
type Marshaller interface {
	String() string
	WriteHeader(header *Header, output io.Writer) error
	WriteSample(sample *Sample, header *Header, output io.Writer) error
}

// Unmarshaller is an interface for reading Samples and Headers from byte streams.
// The byte streams can be anything including files, network connections, console output,
// or in-memory byte buffers.
// Reading is split into three parts: reading the header, receiving the bytes for a sample,
// and parsing those bytes into the actual sample. This separation is done for optimization
// purpose, to enable parallel parsing of samples by separating the data reading part
// from the parsing part. One goroutine can continuously call ReadSampleData(), while multiple
// other routines execute ParseSample() in parallel.
type Unmarshaller interface {
	String() string

	// ReadHeader should read and parse an entire reader from the input.
	// The io.EOF error is not a valid return value and treated as an erro.
	ReadHeader(input *bufio.Reader) (*Header, error)

	// ReadSampleData reads data from the stream, until a full Sample has been
	// read. The data is not parsed, the only task if this is to read exactly
	// the right amount of data. The size of the Sample should be known from the header.
	// The io.EOF error is a valid return value and indicates the end of the stream.
	// If io.EOF occurs too early the stream, it should be converted to io.ErrUnexpectedEOF
	// to avoid confusion.
	ReadSampleData(header *Header, input *bufio.Reader) ([]byte, error)

	// ParseSample uses a header and a byte buffer to parse it to a newly
	// allocated Sample instance. The returned error indicates that the
	// data was in the wrong format.
	ParseSample(header *Header, data []byte) (*Sample, error)
}

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
	case CsvSeparator:
		return new(CsvMarshaller), nil
	case binary_separator:
		return new(BinaryMarshaller), nil
	default:
		return nil, errors.New("Failed to auto-detect format of stream starting with: " + string(peeked))
	}
}

// WriteCascade is a helper type for more concise Write code by avoiding error
// checks on every Write() invokation. Multiple Write calls can be cascaded
// without intermediate checks for errors. The tradeoff/overhead are additional
// no-op Write()/WriteStr() calls after an error has occurred (which is the exception).
type WriteCascade struct {
	// Writer must be set before calling Write. It will receive the Write calls.
	Writer io.Writer

	// Err stores the error that occrred in one of the write calls.
	Err error
}

// Write forwards the call to the contained Writer, but only of no error
// has been encountered yet. If an error occurs, it is stored in the Error field.
func (w *WriteCascade) Write(bytes []byte) error {
	if w.Err == nil {
		_, w.Err = w.Writer.Write(bytes)
	}
	return nil
}

// WriteStr calls Write with a []byte representation of the string parameter.
func (w *WriteCascade) WriteStr(str string) error {
	return w.Write([]byte(str))
}

// WriteByte calls Write with the single parameter byte.
func (w *WriteCascade) WriteByte(b byte) error {
	return w.Write([]byte{b})
}
