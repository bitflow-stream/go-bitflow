package metrics

import (
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"
)

// ==================== Metric ====================
type Metric struct {
	Name   string
	index  int
	sample []Value
}

func (metric *Metric) Set(val Value) {
	metric.sample[metric.index] = val
}

// ==================== Collector ====================
type Collector interface {
	Init() error
	Collect(metric *Metric) error
	Update() error
	SupportedMetrics() []string
	SupportsMetric(metric string) bool
}

var collectors = make(map[Collector]bool)

func RegisterCollector(collector Collector) {
	collectors[collector] = true
}

// Must be called after all collectors have been registered through RegisterCollector
func InitCollectors() error {
	for collector, _ := range collectors {
		if err := collector.Init(); err != nil {
			log.Printf("Failed to initialize data collector %T: %v\n", collector, err)
			delete(collectors, collector)
			continue
		}
		// Do one test update
		if err := collector.Update(); err != nil {
			log.Printf("Failed to update data collector %T: %v\n", collector, err)
			delete(collectors, collector)
			continue
		}
	}
	return nil
}

func AllMetrics() []string {
	var all []string
	for collector, _ := range collectors {
		metrics := collector.SupportedMetrics()
		for _, metric := range metrics {
			all = append(all, metric)
		}
	}
	return all
}

func FilterMetrics(all []string, removeRegexes []*regexp.Regexp) (filtered []string) {
	filtered = make([]string, 0, len(all))
	for _, metric := range all {
		matches := false
		for _, regex := range removeRegexes {
			if matches = regex.MatchString(metric); matches {
				break
			}
		}
		if !matches {
			filtered = append(filtered, metric)
		}
	}
	return
}

func CollectorFor(metric string) Collector {
	for collector, _ := range collectors {
		if collector.SupportsMetric(metric) {
			return collector
		}
	}
	return nil
}

func ConstructSample(metrics []string) (Header, []Value, []Collector, error) {
	set := make(map[Collector]bool)

	header := make(Header, len(metrics))
	values := make([]Value, len(metrics))
	for i, metricName := range metrics {
		collector := CollectorFor(metricName)
		if collector == nil {
			return nil, nil, nil, fmt.Errorf("No collector found for", metricName)
		}
		header[i] = metricName
		metric := Metric{metricName, i, values}

		if err := collector.Collect(&metric); err != nil {
			return nil, nil, nil, fmt.Errorf("Error starting collector for %v: %v\n", metricName, err)
		}
		set[collector] = true
	}

	collectors := make([]Collector, 0, len(set))
	for col, _ := range set {
		collectors = append(collectors, col)
	}
	return header, values, collectors, nil
}

func UpdateCollector(wg *sync.WaitGroup, collector Collector, interval time.Duration) {
	defer wg.Done()
	for {
		err := collector.Update()
		if err != nil {
			log.Println("Warning: Collector update failed:", err)
		}
		time.Sleep(interval)
	}
}

func SinkMetrics(wg *sync.WaitGroup, header Header, values []Value, target MetricSink, interval time.Duration) {
	defer wg.Done()
	for {
		if err := target.Header(header); err != nil {
			log.Printf("Warning: Failed to sink header for %v metrics: %v\n", len(header), err)
		} else {
			for {
				sample := Sample{
					time.Now(),
					values,
				}
				if err := target.Sample(sample); err != nil {
					// When a sample fails, try sending the header again
					log.Printf("Warning: Failed to sink %v metrics: %v\n", len(values), err)
					break
				}
				time.Sleep(interval)
			}
		}
		time.Sleep(interval)
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
