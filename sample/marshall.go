package sample

import (
	"bufio"
	"fmt"
	"io"
)

const (
	time_col = "time"
)

func checkFirstCol(col string) error {
	if col != time_col {
		return fmt.Errorf("Unexpected first column %v, expected %v", col, time_col)
	}
	return nil
}

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
