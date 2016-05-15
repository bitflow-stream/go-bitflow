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
	if _, err := writer.Write([]byte(time_col)); err != nil {
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
	if err := checkFirstCol(fields[0]); err != nil {
		return err
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
