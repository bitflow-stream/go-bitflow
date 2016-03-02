package metrics

import (
	"bufio"
	"encoding/binary"
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
	csv_separator      = ","
	csv_separator_rune = ','
	csv_newline        = "\n"
	csv_time_col       = "time"
	csv_date_layout    = "2006-01-02 15:04:05.999999999"
)

type Value float64
type Header []string
type Sample struct {
	Time   time.Time
	Values []Value
}

func (header Header) WriteBinary(writer io.Writer) error {
	for _, name := range header {
		if _, err := writer.Write([]byte(name)); err != nil {
			return err
		}
		if _, err := writer.Write([]byte{0}); err != nil {
			return err
		}
	}
	if _, err := writer.Write([]byte{0}); err != nil {
		return err
	}
	return nil
}

func (header Header) WriteCsv(writer io.Writer) error {
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

func (header *Header) ReadBinary(reader *bufio.Reader) error {
	*header = nil
	for {
		name, err := reader.ReadBytes(0)
		if err != nil {
			return err
		}
		if len(name) == 0 {
			return nil
		}
		*header = append(*header, string(name[:len(name)-1]))
	}
	return nil
}

func readCsvLine(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	return strings.FieldsFunc(line, func(r rune) bool {
		return r == csv_separator_rune
	}), nil
}

func (header *Header) ReadCsv(reader *bufio.Reader) (err error) {
	*header = nil
	var fields []string
	fields, err = readCsvLine(reader)
	if fields[0] != csv_time_col {
		return fmt.Errorf("Unexpected first column %v, expected %v", fields[0], csv_time_col)
	}
	*header = Header(fields[1:])
	return
}

func (sample *Sample) WriteBinary(writer io.Writer) error {
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

func (sample *Sample) WriteCsv(writer io.Writer) error {
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

// TODO numFields is required since we don't have separators in the binary format
func (sample *Sample) ReadBinary(reader *bufio.Reader, numFields int) error {
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
		sample.Values = append(sample.Values, Value(math.Float64frombits(valBits)))
		return nil
	}
	return nil
}

func (sample *Sample) ReadCsv(reader *bufio.Reader) error {
	sample.Time = time.Time{}
	sample.Values = nil

	fields, err := readCsvLine(reader)
	if len(fields) < 1 {
		return fmt.Errorf("Empty CSV line")
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
