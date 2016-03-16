package metrics

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/antongulenko/golib"
)

const (
	timeBytes          = 8
	valBytes           = 8
	binary_separator   = byte('\n')
	csv_separator      = ","
	csv_separator_rune = ','
	csv_newline        = "\n"
	csv_time_col       = "time"
	csv_date_format    = "2006-01-02 15:04:05.999999999"
	text_date_format   = "2006-01-02 15:04:05.999"
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
	// Time as big-endian uint64 nanoseconds since Unix epoch
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
	line, err := reader.ReadString(csv_newline[0])
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
	if _, err := writer.Write([]byte(sample.Time.Format(csv_date_format))); err != nil {
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
	tim, err := time.Parse(csv_date_format, fields[0])
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
const (
	// Defaults will be applied if value is <= 0
	TEXT_MARSHALLER_DEFAULT_SPACING = 3
	TEXT_MARSHALLER_DEFAULT_WIDTH   = 200 // Automatic for stdin.
	TEXT_MARSHALLER_HEADER_CHAR     = '='
)

type TextMarshaller struct {
	lastHeader Header
	TextWidth  int // Ignored if Columns is > 0
	Columns    int // Will be inferred from TextWidth if <= 0
	Spacing    int
}

func (*TextMarshaller) String() string {
	return "text"
}

func (m *TextMarshaller) WriteHeader(header Header, writer io.Writer) error {
	// HACK A Marshaller should be stateless, but this hack is required for printing...
	// This works as long as only one source is marshalled with this
	m.lastHeader = header
	return nil
}

func (m *TextMarshaller) WriteSample(sample Sample, writer io.Writer) error {
	if len(m.lastHeader) != len(sample.Values) {
		return fmt.Errorf("Cannot write text sample of length %v, expected %v", len(sample.Values), len(m.lastHeader))
	}
	header := sample.Time.Format(text_date_format)
	lines := make([]string, 0, len(sample.Values))
	for i, value := range sample.Values {
		line := fmt.Sprintf("%s = %.4f", m.lastHeader[i], value)
		lines = append(lines, line)
	}

	textWidth, columnWidths := m.calculateWidths(lines, writer)
	m.writeHeader(header, textWidth, writer)
	m.writeLines(lines, columnWidths, writer)
	return nil
}

func (m *TextMarshaller) calculateWidths(lines []string, writer io.Writer) (textWidth int, columnWidths []int) {
	spacing := m.Spacing
	if spacing <= 0 {
		spacing = TEXT_MARSHALLER_DEFAULT_SPACING
	}
	if m.Columns > 0 {
		columnWidths = m.columnWidths(lines, m.Columns, spacing)
		for _, width := range columnWidths {
			textWidth += width
		}
	} else {
		if m.TextWidth > 0 {
			textWidth = m.TextWidth
		} else {
			textWidth = m.defaultTextWidth(writer)
		}
		columns := m.numberOfColumns(lines, textWidth, spacing)
		columnWidths = m.columnWidths(lines, columns, spacing)
	}
	return
}

func (m *TextMarshaller) defaultTextWidth(writer io.Writer) int {
	if writer == os.Stdout {
		if size, err := golib.GetTerminalSize(); err != nil {
			log.Println("Failed to get terminal size:", err)
			return 0
		} else {
			return int(size.Col)
		}
	} else {
		return TEXT_MARSHALLER_DEFAULT_WIDTH
	}
}

func (m *TextMarshaller) numberOfColumns(lines []string, textWidth int, spacing int) int {
	columns := len(lines)
	columnCounter := 0
	width := 0
	for _, line := range lines {
		length := len(line) + spacing
		if width+length > textWidth {
			if columns > columnCounter {
				columns = columnCounter
				if columns <= 1 {
					break
				}
			}
			width = length
			columnCounter = 0
		} else {
			width += length
			columnCounter++
		}
	}
	return columns
}

func (m *TextMarshaller) columnWidths(lines []string, columns int, spacing int) (widths []int) {
	widths = make([]int, columns)
	for i, line := range lines {
		col := i % columns
		length := len(line) + spacing
		if widths[col] < length {
			widths[col] = length
		}
	}
	return
}

func (m *TextMarshaller) writeHeader(header string, textWidth int, writer io.Writer) {
	extraSpace := textWidth - len(header)
	if extraSpace >= 4 {
		lineChars := (extraSpace - 2) / 2
		line := make([]byte, lineChars)
		for i := 0; i < lineChars; i++ {
			line[i] = TEXT_MARSHALLER_HEADER_CHAR
		}
		lineStr := string(line)
		fmt.Fprintln(writer, lineStr, header, lineStr)
	} else {
		fmt.Fprintln(writer, header)
	}
}

func (m *TextMarshaller) writeLines(lines []string, widths []int, writer io.Writer) {
	columns := len(widths)
	for i, line := range lines {
		writer.Write([]byte(line))
		col := i % columns
		if col >= columns-1 || i == len(lines)-1 {
			writer.Write([]byte("\n"))
		} else {
			extraSpace := widths[col] - len(line)
			for j := 0; j < extraSpace; j++ {
				writer.Write([]byte(" "))
			}
		}
	}
}

func (*TextMarshaller) ReadHeader(header *Header, reader *bufio.Reader) error {
	return errors.New("Unmarshalling text data is not supported")
}

func (*TextMarshaller) ReadSample(sample *Sample, reader *bufio.Reader, num int) error {
	return errors.New("Unmarshalling text data is not supported")
}
