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

	// String returns a short description of the Unmarshaller.
	String() string

	// Read must inspect the data in the stream and perform exactly one of two tasks:
	// read a header, or read a sample.
	// The Unmarshaller must be able to distinguish between a header and a sample based
	// on the first bytes received from the stream. If the previousHeader parameter is nil,
	// the Unmarshaller must attempt to receive a header, regardless of the stream contents.
	//
	// If a header is read, it is also parsed and a Header instance is allocated.
	// A pointer to the new header is returned, the sampleData byte-slice must be returned as nil.
	//
	// If sample data is read, Read must read data from the stream, until a full Sample has been read.
	// The sample data is not parsed, the ParseSample() method will be invoked separately.
	// The size of the Sample should be known based on the previousHeader parameter.
	// If sample data is read, is must be returned as the sampleData return value, and the Header pointer
	// must be returned as nil.
	//
	// Error handling:
	// The io.EOF error indicates that the read operation was successfull, and the stream was closed
	// immediately after receiving the data.
	// If io.EOF occurs too early the stream, it should be converted to io.ErrUnexpectedEOF
	// to indicate an actual error condition.
	Read(input *bufio.Reader, previousHeader *Header) (newHeader *Header, sampleData []byte, err error)

	// ParseSample uses a header and a byte buffer to parse it to a newly
	// allocated Sample instance. A non-nil error indicates that the
	// data was in the wrong format.
	ParseSample(header *Header, data []byte) (*Sample, error)
}

// BidiMarshaller is a bidirectional marshaller that combines the
// Marshaller and Unmarshaller interfaces.
type BidiMarshaller interface {
	Read(input *bufio.Reader, previousHeader *Header) (newHeader *Header, sampleData []byte, err error)
	ParseSample(header *Header, data []byte) (*Sample, error)
	WriteHeader(header *Header, output io.Writer) error
	WriteSample(sample *Sample, header *Header, output io.Writer) error
}

func checkFirstCol(col string, readErr error) error {
	if unexpected := unexpectedEOF(readErr); unexpected != nil {
		return unexpected
	}
	if col != time_col {
		if len(col) >= 20 {
			col = col[:20] + "..."
		}
		return fmt.Errorf("First column should be %v, but found: %q", time_col, col)
	}
	return nil
}

func unexpectedEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
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
