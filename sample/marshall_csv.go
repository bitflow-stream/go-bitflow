package sample

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	csv_separator      = ","
	csv_separator_rune = ','
	csv_newline        = "\n"
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
	header.Fields = nil
	var fields []string
	fields, eof, err := readCsvLine(reader)
	if err != nil {
		return err
	}
	if len(fields) == 0 && eof {
		return io.EOF
	}
	if err := checkFirstCol(fields[0]); err != nil {
		return err
	}
	header.HasTags = len(fields) >= 2 && fields[1] == tags_col
	start := 1
	if header.HasTags {
		start++
	}
	header.Fields = fields[start:]
	return nil
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

func (*CsvMarshaller) ReadSample(sample *Sample, header *Header, reader *bufio.Reader) error {
	sample.Time = time.Time{}
	sample.Values = nil
	sample.Tags = make(map[string]string)

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

	start := 1
	if header.HasTags {
		if len(fields) < 2 {
			return fmt.Errorf("Sample too short: %v", fields)
		}
		if err := sample.ParseTagString(fields[1]); err != nil {
			return err
		}
		start++
	}

	for _, field := range fields[start:] {
		val, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return err
		}
		sample.Values = append(sample.Values, Value(val))
	}
	return nil
}
