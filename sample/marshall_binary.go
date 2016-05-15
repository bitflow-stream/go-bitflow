package sample

import (
	"bufio"
	"encoding/binary"
	"io"
	"math"
	"time"
)

const (
	timeBytes        = 8
	valBytes         = 8
	binary_separator = byte('\n')
)

type BinaryMarshaller struct {
}

func (*BinaryMarshaller) String() string {
	return "binary"
}

func (*BinaryMarshaller) WriteHeader(header Header, writer io.Writer) error {
	if _, err := writer.Write([]byte(time_col)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte{binary_separator}); err != nil {
		return err
	}
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
	name, err := reader.ReadBytes(binary_separator)
	if err != nil {
		return err
	}
	if err := checkFirstCol(string(name[:len(name)-1])); err != nil {
		return err
	}
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
