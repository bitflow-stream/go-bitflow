package bitflow

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	csv_time_col    = "time"
	tags_col        = "tags"
	binary_time_col = "timB" // Must not collide with csv_time_col, but have same length

	detect_format_peek        = len(csv_time_col)
	illegal_header_characters = string(CsvSeparator) + string(CsvNewline) + string(BinarySeparator)
)

// Marshaller is an interface for converting Samples and Headers into byte streams.
// The byte streams can be anything including files, network connections, console output,
// or in-memory byte buffers.
type Marshaller interface {
	String() string
	WriteHeader(header *Header, withTags bool, output io.Writer) error
	WriteSample(sample *Sample, header *Header, withTags bool, output io.Writer) error
	// some marshallers might be supposed to send a single sample per request3
	ShouldCloseAfterFirstSample() bool
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
	// The io.EOF error can be returned in two cases: 1) the read operation was successful and complete,
	// but the stream ended immediately afterwards, or 2) the stream was already empty. In the second
	// case, both other return values must be nil.
	// If io.EOF occurs in the middle of reading the stream, it must be converted to io.ErrUnexpectedEOF
	// to indicate an actual error condition.
	Read(input *bufio.Reader, previousHeader *UnmarshalledHeader) (newHeader *UnmarshalledHeader, sampleData []byte, err error)

	// ParseSample uses a header and a byte buffer to parse it to a newly
	// allocated Sample instance. The resulting Sample must have a Value slice with at least the capacity
	// of minValueCapacity. A non-nil error indicates that the data was in the wrong format.
	ParseSample(header *UnmarshalledHeader, minValueCapacity int, data []byte) (*Sample, error)
}

// BidiMarshaller is a bidirectional marshaller that combines the
// Marshaller and Unmarshaller interfaces.
type BidiMarshaller interface {
	Read(input *bufio.Reader, previousHeader *UnmarshalledHeader) (newHeader *UnmarshalledHeader, sampleData []byte, err error)
	ParseSample(header *UnmarshalledHeader, minValueCapacity int, data []byte) (*Sample, error)
	WriteHeader(header *Header, withTags bool, output io.Writer) error
	WriteSample(sample *Sample, header *Header, withTags bool, output io.Writer) error
	String() string
}

// UnmarshalledHeader extends a Header by adding a flag that indicated whether the unmarshalled
// samples will contain tags or not. This enables backwards-compatibility for data input without tags.
type UnmarshalledHeader struct {
	Header
	HasTags bool
}

func readUntil(reader *bufio.Reader, delimiter byte) (data []byte, err error) {
	data, err = reader.ReadBytes(delimiter)
	if err == io.EOF {
		if len(data) > 0 && data[len(data)-1] != delimiter {
			err = io.ErrUnexpectedEOF
		}
	} else if len(data) == 0 && err == nil {
		err = errors.New("Bitflow: empty read")
	}
	return
}

func unexpectedEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

func checkFirstField(expected string, found string) error {
	if found != expected {
		if len(found) >= 20 {
			found = found[:20] + "..."
		}
		return fmt.Errorf("First header field should be '%v', but found: %q", expected, found)
	}
	return nil
}

func checkHeaderField(field string) error {
	if field == "" {
		return errors.New("Header fields cannot be empty")
	}
	if strings.ContainsAny(field, illegal_header_characters) {
		return fmt.Errorf("Header field '%s' contains illegal characters", field)
	}
	return nil
}

func detectFormat(input *bufio.Reader) (Unmarshaller, error) {
	peeked, err := input.Peek(detect_format_peek)
	if err == bufio.ErrBufferFull {
		err = errors.New("IO buffer is too small to auto-detect input stream format")
	}
	if err != nil {
		return nil, err
	}
	return DetectFormatFrom(string(peeked))
}

// DetectFormatFrom uses the start of a marshalled header to determine what unmarshaller
// should be used to decode the header and all following samples.
func DetectFormatFrom(start string) (Unmarshaller, error) {
	if len(start) != detect_format_peek {
		return nil, fmt.Errorf("Cannot auto-detect format of stream based on '%v', need %v characters", start, detect_format_peek)
	}

	switch start {
	case csv_time_col:
		return new(CsvMarshaller), nil
	case binary_time_col:
		return new(BinaryMarshaller), nil
	default:
		return nil, errors.New("Failed to auto-detect format of stream starting with: " + start)
	}
}

// WriteCascade is a helper type for more concise Write code by avoiding error
// checks on every Write() invocation. Multiple Write calls can be cascaded
// without intermediate checks for errors. The trade-off/overhead are additional
// no-op Write()/WriteStr() calls after an error has occurred (which is the exception).
type WriteCascade struct {
	// Writer must be set before calling Write. It will receive the Write calls.
	Writer io.Writer

	// Err stores the error that occurred in one of the write calls.
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

// WriteAny uses the fmt package to format he given object directly into the underlying
// writer. The write is only executed, if previous writes have been successful.
func (w *WriteCascade) WriteAny(i interface{}) error {
	if w.Err == nil {
		_, w.Err = fmt.Fprintf(w.Writer, "%v", i)
	}
	return nil
}
