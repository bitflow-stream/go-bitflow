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
	w := WriteCascade{Writer: writer}
	w.WriteStr(time_col)
	w.WriteByte(binary_separator)
	if header.HasTags {
		w.WriteStr(tags_col)
		w.WriteByte(binary_separator)
	}
	for _, name := range header.Fields {
		w.WriteStr(name)
		w.WriteByte(binary_separator)
	}
	w.WriteByte(binary_separator)
	return w.Err
}

func (*BinaryMarshaller) ReadHeader(reader *bufio.Reader) (header Header, err error) {
	name, err := reader.ReadBytes(binary_separator)
	if err != nil {
		return
	}
	if err = checkFirstCol(string(name[:len(name)-1])); err != nil {
		return
	}

	first := true
	for {
		var nameBytes []byte
		nameBytes, err = reader.ReadBytes(binary_separator)
		if err != nil {
			return
		}
		if len(nameBytes) <= 1 {
			return
		}
		name := string(nameBytes[:len(nameBytes)-1])
		if first && name == tags_col {
			header.HasTags = true
		} else {
			header.Fields = append(header.Fields, name)
		}
		first = false
	}
	return
}

func (m *BinaryMarshaller) WriteSample(sample Sample, header Header, writer io.Writer) error {
	// Time as big-endian uint64 nanoseconds since Unix epoch
	tim := make([]byte, timeBytes)
	binary.BigEndian.PutUint64(tim, uint64(sample.Time.UnixNano()))
	if _, err := writer.Write(tim); err != nil {
		return err
	}

	// Tags
	if header.HasTags {
		str := sample.TagString()
		if _, err := writer.Write([]byte(str)); err != nil {
			return err
		}
		if _, err := writer.Write([]byte{binary_separator}); err != nil {
			return err
		}
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

func (*BinaryMarshaller) ReadSample(header Header, reader *bufio.Reader) (sample Sample, err error) {
	// Time
	tim := make([]byte, timeBytes)
	_, err = io.ReadFull(reader, tim)
	if err != nil {
		return
	}
	timeVal := binary.BigEndian.Uint64(tim)
	sample.Time = time.Unix(0, int64(timeVal))

	// Tags
	if header.HasTags {
		var tags string
		if tags, err = reader.ReadString(binary_separator); err != nil {
			return
		}
		if err = sample.ParseTagString(tags); err != nil {
			return
		}
	}

	// Values
	for i := 0; i < len(header.Fields); i++ {
		val := make([]byte, valBytes)
		if _, err = io.ReadFull(reader, val); err != nil {
			return
		}
		valBits := binary.BigEndian.Uint64(val)
		value := math.Float64frombits(valBits)
		sample.Values = append(sample.Values, Value(value))
	}
	return
}
