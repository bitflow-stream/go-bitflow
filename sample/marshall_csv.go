package sample

import (
	"bufio"
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
	CsvSeparator = ','

	// CsvNewline is used by CsvMarshaller after outputting the header line and
	// each sample.
	CsvNewline = "\n"

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
// representation of the timestamp (see CsvDateFormat), then a space-separated
// key-value list for the tags (only if the 'tags' field was included in the header),
// and then all the metric values in the same order as on the preceding header line.
// To follow the semantics of a correct CSV file, every changing header should start
// a new CSV file.
//
// There are no configuration options for CsvMarshaller.
type CsvMarshaller struct {
}

// String implements the Marshaller interface.
func (*CsvMarshaller) String() string {
	return "CSV"
}

// WriteHeader implements the Marshaller interface by printing a CSV header line.
func (*CsvMarshaller) WriteHeader(header *Header, writer io.Writer) error {
	w := WriteCascade{Writer: writer}
	w.WriteStr(time_col)
	if header.HasTags {
		w.WriteByte(CsvSeparator)
		w.WriteStr(tags_col)
	}
	for _, name := range header.Fields {
		w.WriteByte(CsvSeparator)
		w.WriteStr(name)
	}
	w.WriteStr(CsvNewline)
	return w.Err
}

func splitCsvLine(line []byte) []string {
	return strings.Split(string(line), string(CsvSeparator))
}

// ReadHeader implements the Unmarshaller interface by reading and parsing a CSV header line.
func (*CsvMarshaller) ReadHeader(reader *bufio.Reader) (header *Header, err error) {
	line, err := reader.ReadBytes(CsvNewline[0])
	if err == io.EOF {
		err = nil
	} else if err != nil {
		return
	} else if len(line) > 0 {
		line = line[:len(line)-1] // Strip newline char
	}
	if len(line) == 0 {
		err = errors.New("Empty header")
		return
	}
	fields := splitCsvLine(line)
	if err = checkFirstCol(fields[0]); err != nil {
		return
	}
	header = new(Header)
	header.HasTags = len(fields) >= 2 && fields[1] == tags_col
	start := 1
	if header.HasTags {
		start++
	}
	header.Fields = fields[start:]
	return
}

// WriteSample implements the Marshaller interface by writing a CSV line.
func (*CsvMarshaller) WriteSample(sample *Sample, header *Header, writer io.Writer) error {
	w := WriteCascade{Writer: writer}
	w.WriteStr(sample.Time.Format(CsvDateFormat))
	if header.HasTags {
		tags := sample.TagString()
		w.WriteByte(CsvSeparator)
		w.WriteStr(tags)
	}
	for _, value := range sample.Values {
		w.WriteByte(CsvSeparator)
		w.WriteStr(fmt.Sprintf("%v", value))
	}
	w.WriteStr(CsvNewline)
	return w.Err
}

// ReadSampleData implements the Unmarshaller interface by reading data until
// the next newline character.
func (*CsvMarshaller) ReadSampleData(header *Header, input *bufio.Reader) ([]byte, error) {
	data, err := input.ReadBytes(CsvNewline[0])
	if err == io.EOF {
		if len(data) > 0 {
			err = nil
		}
	} else if len(data) > 0 {
		data = data[:len(data)-1] // Strip newline char
	}
	return data, err
}

// ParseSample implements the Unmarshaller interface by parsing a CSV line.
func (*CsvMarshaller) ParseSample(header *Header, data []byte) (sample *Sample, err error) {
	fields := splitCsvLine(data)
	sample = new(Sample)
	sample.Time, err = time.Parse(CsvDateFormat, fields[0])
	if err != nil {
		return
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
