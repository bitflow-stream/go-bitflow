package metrics

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"time"
)

const (
	valBytes = 8
)

type Value float64

type Metric struct {
	Tag string
	Val Value
}

func (metric *Metric) WriteTo(writer io.Writer) error {
	if metric == nil {
		return errors.New("Cannot marshal nil *Metric")
	}
	if _, err := writer.Write([]byte(metric.Tag)); err != nil {
		return err
	}
	if _, err := writer.Write([]byte("\x00")); err != nil {
		return err
	}
	valBits := math.Float64bits(float64(metric.Val))
	var val [valBytes]byte
	binary.BigEndian.PutUint64(val[:], valBits)
	if _, err := writer.Write(val[:]); err != nil {
		return err
	}
	return nil
}

func (metric *Metric) ReadFrom(reader *bufio.Reader) error {
	if metric == nil {
		return errors.New("Cannot unmarshal into nil *Metric")
	}
	tag, err := reader.ReadBytes('\x00')
	if err != nil {
		return err
	}
	metric.Tag = string(tag[:len(tag)-1])
	var val [valBytes]byte
	_, err = io.ReadFull(reader, val[:])
	if err != nil {
		return err
	}
	valBits := binary.BigEndian.Uint64(val[:])
	metric.Val = Value(math.Float64frombits(valBits))
	return nil
}

func (metric *Metric) String() string {
	if metric == nil {
		return "<nil> Metric"
	}
	return fmt.Sprintf("%v: %v", metric.Tag, metric.Val)
}

func (val Value) DiffValue(logback LogbackValue, interval time.Duration) Value {
	switch previous := logback.(type) {
	case Value:
		return Value(val-previous) / Value(interval.Seconds())
	case *Value:
		return Value(val-*previous) / Value(interval.Seconds())
	default:
		log.Printf("Error: Cannot diff %v (%T) and %v (%T)\n", val, val, logback, logback)
		return Value(0)
	}
}

// ================================= Ring logback of recorded Values =================================
type ValueRing struct {
	values []TimedValue
	head   int // actually head+1
}

func NewValueRing(length int) ValueRing {
	return ValueRing{
		values: make([]TimedValue, length),
	}
}

type LogbackValue interface {
	DiffValue(previousValue LogbackValue, interval time.Duration) Value
}

type TimedValue struct {
	time.Time // Timestamp of recording
	val       LogbackValue
}

func (ring *ValueRing) Add(val LogbackValue) {
	ring.values[ring.head] = TimedValue{time.Now(), val}
	if ring.head >= len(ring.values)-1 {
		ring.head = 0
	} else {
		ring.head++
	}
}

func (ring *ValueRing) getHead() TimedValue {
	headIndex := ring.head
	if headIndex <= 0 {
		headIndex = len(ring.values) - 1
	} else {
		headIndex--
	}
	return ring.values[headIndex]
}

// Does not check for empty ring
func (ring *ValueRing) get(before time.Time) (result TimedValue) {
	walkRing := func(i int) bool {
		if ring.values[i].val == nil {
			return false
		}
		result = ring.values[i]
		if result.Time.Before(before) {
			return false
		}
		return true
	}
	for i := ring.head - 1; i >= 0; i-- {
		if !walkRing(i) {
			return
		}
	}
	for i := len(ring.values) - 1; i >= ring.head; i-- {
		if !walkRing(i) {
			return
		}
	}
	return
}

func (ring *ValueRing) GetDiff(before time.Duration) Value {
	head := ring.getHead()
	beforeTime := head.Time.Add(-before)
	previous := ring.get(beforeTime)
	return head.val.DiffValue(previous.val, head.Time.Sub(previous.Time))
}
