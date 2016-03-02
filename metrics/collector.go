package metrics

import (
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"
)

type Metric struct {
	Name string
	Val  Value
	Time time.Time
}

type Metrics []*Metric

type Collector interface {
	Init() error
	Collect(metric *Metric) error
	Update() error
	SupportedMetrics() []string
	SupportsMetric(metric string) bool
}

var collectors []Collector

func RegisterCollector(collector Collector) {
	collectors = append(collectors, collector)
}

// Must be called after all collectors have been registered through RegisterCollector
func InitCollectors() error {
	for _, collector := range collectors {
		if err := collector.Init(); err != nil {
			return err
		}
	}
	return nil
}

func (metric *Metric) String() string {
	if metric == nil {
		return "<nil> Metric"
	}
	timeStr := metric.Time.Format("2006-01-02 15:04:05.999")
	return fmt.Sprintf("(%v) %v = %.4f", timeStr, metric.Name, metric.Val)
}

func (metrics Metrics) Header() (header Header) {
	header = make(Header, 0, len(metrics))
	for _, metric := range metrics {
		header = append(header, metric.Name)
	}
	return
}

func (metrics Metrics) Sample() Sample {
	// Use the timestamp of the first metric.
	// TODO refactor
	if len(metrics) == 0 {
		return Sample{}
	}
	values := make([]Value, len(metrics))
	for _, metric := range metrics {
		values = append(values, metric.Val)
	}
	return Sample{
		Time:   metrics[0].Time,
		Values: values,
	}
}

func AllMetrics() []*Metric {
	var all []*Metric
	for _, collector := range collectors {
		metrics := collector.SupportedMetrics()
		for _, metric := range metrics {
			all = append(all, &Metric{
				Name: metric,
			})
		}
	}
	return all
}

func CollectorFor(metric *Metric) Collector {
	for _, collector := range collectors {
		if collector.SupportsMetric(metric.Name) {
			return collector
		}
	}
	return nil
}

func CollectMetrics(metrics ...*Metric) ([]Collector, error) {
	set := make(map[Collector]bool)

	for _, metric := range metrics {
		collector := CollectorFor(metric)
		if collector == nil {
			return nil, fmt.Errorf("Warning: no collector found for", metric.Name)
		}
		if err := collector.Collect(metric); err != nil {
			return nil, fmt.Errorf("Error starting collector for %v: %v\n", metric.Name, err)
		}
		set[collector] = true
	}

	result := make([]Collector, 0, len(set))
	for col, _ := range set {
		result = append(result, col)
	}
	return result, nil
}

func FilterMetrics(all []*Metric, removeRegexes []*regexp.Regexp) (filtered []*Metric) {
	filtered = make([]*Metric, 0, len(all))
	for _, metric := range all {
		matches := false
		for _, regex := range removeRegexes {
			if matches = regex.MatchString(metric.Name); matches {
				break
			}
		}
		if !matches {
			filtered = append(filtered, metric)
		}
	}
	return
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

func SinkMetrics(wg *sync.WaitGroup, metrics Metrics, target MetricSink, interval time.Duration) {
	defer wg.Done()
	for {
		target.Header(metrics.Header())
		err := target.Sample(metrics.Sample())
		if err != nil {
			log.Printf("Warning: Failed to sink %v metrics\n", len(metrics), err)
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
