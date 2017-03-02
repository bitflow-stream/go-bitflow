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

	log "github.com/Sirupsen/logrus"
)

var WarnObsoleteBinaryFormat = true

const (
	// CsvSeparator is the character separating fields in the marshalled output
	// of CsvMarshaller.
	CsvSeparator = ','

	// CsvNewline is used by CsvMarshaller after outputting the header line and
	// each sample.
	CsvNewline = '\n'

	// CsvDateFormat is the format used by CsvMarshaller to marshall the timestamp
	// of samples.
	CsvDateFormat = "2006-01-02 15:04:05.999999999"
)

// CsvMarshaller marshalls Headers and Samples to a CSV format.
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
}

// String implements the Marshaller interface.
func (CsvMarshaller) String() string {
	return "CSV"
}

// WriteHeader implements the Marshaller interface by printing a CSV header line.
func (CsvMarshaller) WriteHeader(header *Header, writer io.Writer) error {
	w := WriteCascade{Writer: writer}
	w.WriteStr(csv_time_col)
	if header.HasTags {
		w.WriteByte(CsvSeparator)
		w.WriteStr(tags_col)
	}
	for _, name := range header.Fields {
		if err := checkHeaderField(name); err != nil {
			return err
		}
		w.WriteByte(CsvSeparator)
		w.WriteStr(name)
	}
	w.WriteStr(string(CsvNewline))
	return w.Err
}

// WriteSample implements the Marshaller interface by writing a CSV line.
func (CsvMarshaller) WriteSample(sample *Sample, header *Header, writer io.Writer) error {
	w := WriteCascade{Writer: writer}
	w.WriteStr(sample.Time.UTC().Format(CsvDateFormat))
	if header.HasTags {
		tags := sample.TagString()
		w.WriteByte(CsvSeparator)
		w.WriteStr(tags)
	}
	for _, value := range sample.Values {
		w.WriteByte(CsvSeparator)
		w.WriteAny(value)
	}
	w.WriteStr(string(CsvNewline))
	return w.Err
}

func splitCsvLine(line []byte) []string {
	return strings.Split(string(line), string(CsvSeparator))
}

// Read implements the Unmarshaller interface by reading CSV line from the input stream.
// Based on the first field, Read decides whether the line represents a header or a Sample.
// In case of a header, the CSV fields are split and parsed to a Header instance.
// In case of a Sample, the data for the line is returned unparsed.
func (c CsvMarshaller) Read(reader *bufio.Reader, previousHeader *Header) (*Header, []byte, error) {
	line, err := readUntil(reader, CsvNewline)
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
	index := bytes.Index(line, []byte{CsvSeparator})
	var firstField string
	if index < 0 {
		firstField = string(line) // Only one field
	} else {
		firstField = string(line[:index])
	}

	switch {
	case previousHeader == nil:
		if checkErr := checkFirstField(csv_time_col, firstField); checkErr != nil {
			return nil, nil, checkErr
		}
		return c.parseHeader(line), nil, err
	case firstField == csv_time_col:
		return c.parseHeader(line), nil, err
	default:
		return nil, line, err
	}
}

func (CsvMarshaller) parseHeader(line []byte) *Header {
	fields := splitCsvLine(line)
	if WarnObsoleteBinaryFormat && len(fields) == 1 {
		log.Warnln("CSV header contains only time field. This might be the old binary format, " +
			"use the 'old_binary_format' tag from the go-bitflow-pipeline repository.")
	}
	hasTags := len(fields) >= 2 && fields[1] == tags_col
	header := &Header{HasTags: hasTags}
	start := 1
	if hasTags {
		start++
	}
	header.Fields = fields[start:]
	if len(header.Fields) == 0 {
		header.Fields = nil
	}
	return header
}

// ParseSample implements the Unmarshaller interface by parsing a CSV line.
func (CsvMarshaller) ParseSample(header *Header, minValueCapacity int, data []byte) (sample *Sample, err error) {
	fields := splitCsvLine(data)
	var t time.Time
	t, err = time.Parse(CsvDateFormat, fields[0])
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
		if err = sample.ParseTagString(fields[1]); err != nil {
			return
		}
		start++
	}

	for _, field := range fields[start:] {
		var val float64
		if val, err = strconv.ParseFloat(field, 64); err != nil {
			return
		}
		sample.Values = append(sample.Values, Value(val))
	}
	return sample, nil
}
