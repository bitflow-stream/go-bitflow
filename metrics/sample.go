package metrics

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	timeBytes          = 8
	valBytes           = 8
	binary_separator   = byte('\n')
	csv_separator      = ","
	csv_separator_rune = ','
	csv_newline        = "\n"
	csv_time_col       = "time"
	csv_date_layout    = "2006-01-02 15:04:05.999999999"
)

type Marshaller interface {
	String() string
	WriteHeader(header Header, output io.Writer) error
	WriteSample(sample Sample, output io.Writer) error
}

type Unmarshaller interface {
	String() string
	ReadHeader(header *Header, input *bufio.Reader) error
	ReadSample(sample *Sample, input *bufio.Reader, num int) error
}

type MetricMarshaller interface {
	String() string
	WriteHeader(header Header, output io.Writer) error
	WriteSample(sample Sample, output io.Writer) error
	ReadHeader(header *Header, input *bufio.Reader) error
	ReadSample(sample *Sample, input *bufio.Reader, num int) error
}

type Value float64
type Header []string
type Sample struct {
	Time   time.Time
	Values []Value
}

// ==================== Binary Format ====================
type BinaryMarshaller struct {
}

func (*BinaryMarshaller) String() string {
	return "binary"
}

func (*BinaryMarshaller) WriteHeader(header Header, writer io.Writer) error {
	for _, name := range header {
		if _, err := writer.Write([]byte(name)); err != nil {
			return err
		}
		if _, err := writer.Write([]byte{binary_separator}); err != nil {
			return err
		}
	}
	if _, err := writer.Write([]byte{binary_separator}); err != nil {
		return err
	}
	return nil
}

func (*BinaryMarshaller) ReadHeader(header *Header, reader *bufio.Reader) error {
	*header = nil
	for {
		name, err := reader.ReadBytes(binary_separator)
		if err != nil {
			return err
		}
		if len(name) <= 1 {
			return nil
		}
		*header = append(*header, string(name[:len(name)-1]))
	}
	return nil
}

func (*BinaryMarshaller) WriteSample(sample Sample, writer io.Writer) error {
	// Time as uint64 nanoseconds since Unix epoch
	tim := make([]byte, timeBytes)
	binary.BigEndian.PutUint64(tim, uint64(sample.Time.UnixNano()))
	if _, err := writer.Write(tim); err != nil {
		return err
	}

	// Values as big-endian double precision
	for _, value := range sample.Values {
		valBits := math.Float64bits(float64(value))
		val := make([]byte, valBytes)
		binary.BigEndian.PutUint64(val, valBits)
		if _, err := writer.Write(val); err != nil {
			return err
		}
	}
	return nil
}

func (*BinaryMarshaller) ReadSample(sample *Sample, reader *bufio.Reader, numFields int) error {
	sample.Time = time.Time{}
	sample.Values = nil

	// Time
	tim := make([]byte, timeBytes)
	_, err := io.ReadFull(reader, tim)
	if err != nil {
		return err
	}
	timeVal := binary.BigEndian.Uint64(tim)
	sample.Time = time.Unix(0, int64(timeVal))

	// Values
	for i := 0; i < numFields; i++ {
		val := make([]byte, valBytes)
		_, err = io.ReadFull(reader, val)
		if err != nil {
			return err
		}
		valBits := binary.BigEndian.Uint64(val)
		value := math.Float64frombits(valBits)
		sample.Values = append(sample.Values, Value(value))
	}
	return nil
}

// ==================== CSV Format ====================
type CsvMarshaller struct {
}

func (*CsvMarshaller) String() string {
	return "CSV"
}

func (*CsvMarshaller) WriteHeader(header Header, writer io.Writer) error {
	if _, err := writer.Write([]byte(csv_time_col)); err != nil {
		return err
	}
	for _, name := range header {
		if _, err := writer.Write([]byte(csv_separator)); err != nil {
			return err
		}
		if _, err := writer.Write([]byte(name)); err != nil {
			return err
		}
	}
	_, err := writer.Write([]byte(csv_newline))
	if err != nil {
		return err
	}
	return nil
}

func readCsvLine(reader *bufio.Reader) ([]string, bool, error) {
	line, err := reader.ReadString('\n')
	eof := err == io.EOF
	if err != nil && !eof {
		return nil, false, err
	}
	if len(line) == 0 {
		return nil, eof, nil
	}
	line = line[:len(line)-1] // Strip newline char
	return strings.FieldsFunc(line, func(r rune) bool {
		return r == csv_separator_rune
	}), eof, nil
}

func (*CsvMarshaller) ReadHeader(header *Header, reader *bufio.Reader) error {
	*header = nil
	var fields []string
	fields, eof, err := readCsvLine(reader)
	if err != nil {
		return err
	}
	if len(fields) == 0 && eof {
		return io.EOF
	}
	if fields[0] != csv_time_col {
		return fmt.Errorf("Unexpected first column %v, expected %v", fields[0], csv_time_col)
	}
	*header = Header(fields[1:])
	return nil
}

func (*CsvMarshaller) WriteSample(sample Sample, writer io.Writer) error {
	if _, err := writer.Write([]byte(sample.Time.Format(csv_date_layout))); err != nil {
		return err
	}
	for _, value := range sample.Values {
		if _, err := writer.Write([]byte(csv_separator)); err != nil {
			return err
		}
		if _, err := writer.Write([]byte(fmt.Sprintf("%v", value))); err != nil {
			return err
		}
	}
	_, err := writer.Write([]byte(csv_newline))
	if err != nil {
		return err
	}
	return nil
}

func (*CsvMarshaller) ReadSample(sample *Sample, reader *bufio.Reader, num int) error {
	sample.Time = time.Time{}
	sample.Values = nil

	fields, eof, err := readCsvLine(reader)
	if err != nil {
		return err
	}
	if len(fields) == 0 && eof {
		return io.EOF
	}
	tim, err := time.Parse(csv_date_layout, fields[0])
	if err != nil {
		return err
	}
	sample.Time = tim
	for _, field := range fields[1:] {
		val, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return err
		}
		sample.Values = append(sample.Values, Value(val))
	}
	return nil
}

// ==================== Text Format ====================
type TextMarshaller struct {
	lastHeader Header
}

func (*TextMarshaller) String() string {
	return "text"
}

func (m *TextMarshaller) WriteHeader(header Header, writer io.Writer) error {
	// TODO A Marshaller should be stateless, but this hack is required for printing...
	// This works as long as only one source is marshalled with this
	m.lastHeader = header
	return nil
}

func (m *TextMarshaller) WriteSample(sample Sample, writer io.Writer) error {
	if len(m.lastHeader) != len(sample.Values) {
		return fmt.Errorf("Canot write text sample of length %v, expected %v", len(sample.Values), len(m.lastHeader))
	}
	timeStr := sample.Time.Format("2006-01-02 15:04:05.999")
	fmt.Fprintf(writer, "%s: ", timeStr)
	for i, value := range sample.Values {
		fmt.Fprintf(writer, "%s = %.4f", m.lastHeader[i], value)
		if i < len(sample.Values)-1 {
			fmt.Fprintf(writer, ", ")
		}
	}
	fmt.Fprintln(writer)
	return nil
}

func (*TextMarshaller) ReadHeader(header *Header, reader *bufio.Reader) error {
	return errors.New("Unmarshalling text data is not supported")
}

func (*TextMarshaller) ReadSample(sample *Sample, reader *bufio.Reader, num int) error {
	return errors.New("Unmarshalling text data is not supported")
}
