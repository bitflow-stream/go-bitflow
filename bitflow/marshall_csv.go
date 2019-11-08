package bitflow

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	// CsvSeparator is the character separating fields in the marshalled output
	// of CsvMarshaller.
	DefaultCsvSeparator = ','

	// CsvNewline is used by CsvMarshaller after outputting the header line and
	// each sample.
	DefaultCsvNewline = '\n'

	// CsvDateFormat is the format used by CsvMarshaller to marshall the timestamp
	// of samples.
	DefaultCsvDateFormat = "2006-01-02 15:04:05.999999999"
)

// CsvMarshaller marshals Headers and Samples to a CSV format.
//
// Every header is marshalled as a comma-separated CSV header line.
// The first field is 'time', the second field is 'tags' (if the following samples
// contain tags). After that the header contains a list of all metrics.
//
// Every sample is marshalled to a comma-separated line starting with a textual
// representation of the timestamp (see CsvDateFormat, UTC timezone), then a space-separated
// key-value list for the tags (only if the 'tags' field was included in the header),
// and then all the metric values in the same order as on the preceding header line.
// To follow the semantics of a correct CSV file, every changed header should start
// a new CSV file.
//
// Every CSV line must be terminated by a newline character (including the last line in a file).
//
// CsvMarshaller can deal with multiple header declarations in the same file or
// data stream. A line that begins with the string "time" is assumed to start a new header,
// since samples usually start with a timestamp, which cannot be formatted as "time".
//
// There are no configuration options for CsvMarshaller.
type CsvMarshaller struct {
	ValueSeparator        byte
	LineSeparator         byte
	DateFormat            string
	TimeColumn            string
	TagsColumn            string
	UnmarshallSkipColumns uint
}

func (c *CsvMarshaller) valSep() byte {
	if c.ValueSeparator == 0 {
		return DefaultCsvSeparator
	}
	return c.ValueSeparator
}

func (c *CsvMarshaller) lineSep() byte {
	if c.LineSeparator == 0 {
		return DefaultCsvNewline
	}
	return c.LineSeparator
}

func (c *CsvMarshaller) dateFormat() string {
	if c.DateFormat == "" {
		return DefaultCsvDateFormat
	}
	return c.DateFormat
}

func (c *CsvMarshaller) timeCol() string {
	if c.TimeColumn == "" {
		return DefaultCsvTimeColumn
	}
	return c.TimeColumn
}

func (c *CsvMarshaller) tagsCol() string {
	if c.TagsColumn == "" {
		return TagsColumn
	}
	return c.TagsColumn
}

// ShouldCloseAfterFirstSample defines that csv streams can stream without closing
func (CsvMarshaller) ShouldCloseAfterFirstSample() bool {
	return false
}

// String implements the Marshaller interface.
func (CsvMarshaller) String() string {
	return "CSV"
}

// WriteHeader implements the Marshaller interface by printing a CSV header line.
func (c *CsvMarshaller) WriteHeader(header *Header, withTags bool, writer io.Writer) error {
	w := WriteCascade{Writer: writer}
	w.WriteStr(c.timeCol())
	if withTags {
		w.WriteByte(c.valSep())
		w.WriteStr(c.tagsCol())
	}
	for _, name := range header.Fields {
		if err := checkHeaderField(name); err != nil {
			return err
		}
		w.WriteByte(c.valSep())
		w.WriteStr(name)
	}
	w.WriteStr(string(c.lineSep()))
	return w.Err
}

// WriteSample implements the Marshaller interface by writing a CSV line.
func (c *CsvMarshaller) WriteSample(sample *Sample, header *Header, withTags bool, writer io.Writer) error {
	w := WriteCascade{Writer: writer}
	w.WriteStr(sample.Time.UTC().Format(c.dateFormat()))
	if withTags {
		tags := sample.TagString()
		w.WriteByte(c.valSep())
		w.WriteStr(tags)
	}
	for _, value := range sample.Values {
		w.WriteByte(c.valSep())
		w.WriteAny(value)
	}
	w.WriteStr(string(c.lineSep()))
	return w.Err
}

func (c *CsvMarshaller) splitCsvLine(line []byte) []string {
	return strings.Split(string(line), string(c.valSep()))
}

// Read implements the Unmarshaller interface by reading CSV line from the input stream.
// Based on the first field, Read decides whether the line represents a header or a Sample.
// In case of a header, the CSV fields are split and parsed to a Header instance.
// In case of a Sample, the data for the line is returned without parsing it.
func (c *CsvMarshaller) Read(reader *bufio.Reader, previousHeader *UnmarshalledHeader) (*UnmarshalledHeader, []byte, error) {
	line, err := readUntil(reader, c.lineSep())
	if err == io.EOF {
		if len(line) == 0 {
			return nil, nil, err
		} else {
			// Ignore here
		}
	} else if err != nil {
		return nil, nil, err
	} else if len(line) == 1 {
		return nil, nil, errors.New("Empty CSV line")
	} else if len(line) > 0 {
		line = line[:len(line)-1] // Strip newline char
	}
	index := bytes.Index(line, []byte{c.valSep()})
	var firstField string
	if index < 0 {
		firstField = string(line) // Only one field
	} else {
		firstField = string(line[:index])
	}

	switch {
	case previousHeader == nil:
		if checkErr := checkFirstField(c.timeCol(), firstField); checkErr != nil {
			return nil, nil, checkErr
		}
		return c.parseHeader(line), nil, err
	case firstField == c.timeCol():
		return c.parseHeader(line), nil, err
	default:
		return nil, line, err
	}
}

func (c *CsvMarshaller) skipFields(fields []string) []string {
	skip := int(c.UnmarshallSkipColumns)
	if skip >= len(fields) {
		return nil
	} else {
		return fields[skip:]
	}
}

func (c *CsvMarshaller) parseHeader(line []byte) *UnmarshalledHeader {
	fields := c.splitCsvLine(line)
	hasTags := len(fields) >= 2 && fields[1] == c.tagsCol()
	header := &UnmarshalledHeader{
		HasTags: hasTags,
	}
	start := 1
	if hasTags {
		start++
	}
	fields = c.skipFields(fields[start:])
	if len(fields) == 0 {
		fields = nil
	}
	header.Fields = fields
	return header
}

// ParseSample implements the Unmarshaller interface by parsing a CSV line.
func (c *CsvMarshaller) ParseSample(header *UnmarshalledHeader, minValueCapacity int, data []byte) (sample *Sample, err error) {
	fields := c.splitCsvLine(data)
	var t time.Time
	t, err = time.Parse(c.dateFormat(), fields[0])
	if err != nil {
		return
	}
	var values []Value
	if minValueCapacity > 0 {
		values = make([]Value, 0, minValueCapacity)
	}
	sample = &Sample{
		Values: values,
		Time:   t,
	}

	start := 1
	if header.HasTags {
		if len(fields) < 2 {
			err = fmt.Errorf("Sample too short: %v", fields)
			return
		}
		sample.ParseTagString(fields[1])
		start++
	}

	fields = c.skipFields(fields[start:])
	for _, field := range fields {
		var val float64
		if val, err = strconv.ParseFloat(field, 64); err != nil {
			return
		}
		sample.Values = append(sample.Values, Value(val))
	}
	return sample, nil
}
