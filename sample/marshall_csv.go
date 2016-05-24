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
	csv_separator_rune = ','
	csv_newline        = "\n"
	csv_separator      = string(csv_separator_rune)
	csv_date_format    = "2006-01-02 15:04:05.999999999"
)

type CsvMarshaller struct {
}

func (*CsvMarshaller) String() string {
	return "CSV"
}

func (*CsvMarshaller) WriteHeader(header Header, writer io.Writer) error {
	w := WriteCascade{Writer: writer}
	w.WriteStr(time_col)
	if header.HasTags {
		w.WriteStr(csv_separator)
		w.WriteStr(tags_col)
	}
	for _, name := range header.Fields {
		w.WriteStr(csv_separator)
		w.WriteStr(name)
	}
	w.WriteStr(csv_newline)
	return w.Err
}

func splitCsvLine(line []byte) []string {
	return strings.FieldsFunc(string(line), func(r rune) bool {
		return r == csv_separator_rune
	})
}

func (*CsvMarshaller) ReadHeader(reader *bufio.Reader) (header Header, err error) {
	line, err := reader.ReadBytes(csv_newline[0])
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
	header.HasTags = len(fields) >= 2 && fields[1] == tags_col
	start := 1
	if header.HasTags {
		start++
	}
	header.Fields = fields[start:]
	return
}

func (*CsvMarshaller) WriteSample(sample Sample, header Header, writer io.Writer) error {
	w := WriteCascade{Writer: writer}
	w.WriteStr(sample.Time.Format(csv_date_format))
	if header.HasTags {
		tags := sample.TagString()
		w.WriteStr(csv_separator)
		w.WriteStr(tags)
	}
	for _, value := range sample.Values {
		w.WriteStr(csv_separator)
		w.WriteStr(fmt.Sprintf("%v", value))
	}
	w.WriteStr(csv_newline)
	return w.Err
}

func (*CsvMarshaller) ReadSampleData(header Header, input *bufio.Reader) ([]byte, error) {
	data, err := input.ReadBytes(csv_newline[0])
	if err == io.EOF {
		if len(data) > 0 {
			err = nil
		}
	} else if len(data) > 0 {
		data = data[:len(data)-1] // Strip newline char
	}
	return data, err
}

func (*CsvMarshaller) ParseSample(header Header, data []byte) (sample Sample, err error) {
	fields := splitCsvLine(data)
	sample.Time, err = time.Parse(csv_date_format, fields[0])
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
