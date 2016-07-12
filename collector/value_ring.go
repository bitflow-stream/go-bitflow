package collector

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/data2go/sample"
)

type ValueRing struct {
	interval time.Duration // Can be set to use GetDefaultDiff()

	values []TimedValue
	head   int // actually head+1

	aggregator   LogbackValue
	previousDiff sample.Value
}

func NewValueRing(length int, interval time.Duration) ValueRing {
	return ValueRing{
		values:   make([]TimedValue, length),
		interval: interval,
	}
}

type LogbackValue interface {
	DiffValue(previousValue LogbackValue, interval time.Duration) sample.Value
	AddValue(val LogbackValue) LogbackValue
}

type TimedValue struct {
	time.Time // Timestamp of recording
	val       LogbackValue
}

func (ring *ValueRing) AddToHead(val LogbackValue) {
	if ring.aggregator == nil {
		ring.aggregator = val
	} else {
		ring.aggregator = ring.aggregator.AddValue(val)
	}
}

func (ring *ValueRing) FlushHead() {
	ring.values[ring.head] = TimedValue{time.Now(), ring.aggregator}
	if ring.head >= len(ring.values)-1 {
		ring.head = 0
	} else {
		ring.head++
	}
	ring.aggregator = nil
}

func (ring *ValueRing) Add(val LogbackValue) {
	ring.AddToHead(val)
	ring.FlushHead()
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

func (ring *ValueRing) flush(start int) {
	// Flush all older values, starting (including) the start
	if start < 0 {
		start += len(ring.values)
	}
	for i := start; i >= 0; i-- {
		if ring.values[i].val == nil {
			return
		}
		ring.values[i].val = nil
	}
	for i := len(ring.values) - 1; i >= ring.head; i-- {
		if ring.values[i].val == nil {
			return
		}
		ring.values[i].val = nil
	}
}

func (ring *ValueRing) GetDiff() sample.Value {
	val := ring.getDiffInterval(ring.interval)
	if val < 0 {
		// Likely means a number has overflown. Temporarily stick to same value.
		val = ring.previousDiff
		ring.flush(ring.head - 2) // Only keep the latest sample
	} else {
		ring.previousDiff = val
	}
	return val
}

func (ring *ValueRing) getDiffInterval(before time.Duration) sample.Value {
	head := ring.getHead()
	if head.val == nil {
		// Probably empty ring
		return sample.Value(0)
	}
	beforeTime := head.Time.Add(-before)
	previous := ring.get(beforeTime)
	if previous.val == nil {
		return sample.Value(0)
	}
	interval := head.Time.Sub(previous.Time)
	if interval == 0 {
		return sample.Value(0)
	}
	return head.val.DiffValue(previous.val, interval)
}

type StoredValue sample.Value

func (val StoredValue) DiffValue(logback LogbackValue, interval time.Duration) sample.Value {
	switch previous := logback.(type) {
	case StoredValue:
		return sample.Value(val-previous) / sample.Value(interval.Seconds())
	case *StoredValue:
		return sample.Value(val-*previous) / sample.Value(interval.Seconds())
	default:
		log.Errorf("Cannot diff %v (%T) and %v (%T)", val, val, logback, logback)
		return sample.Value(0)
	}
}

func (val StoredValue) AddValue(incoming LogbackValue) LogbackValue {
	switch other := incoming.(type) {
	case StoredValue:
		return StoredValue(val + other)
	case *StoredValue:
		return StoredValue(val + *other)
	default:
		log.Errorf("Cannot add %v (%T) and %v (%T)", val, val, incoming, incoming)
		return StoredValue(0)
	}
}
