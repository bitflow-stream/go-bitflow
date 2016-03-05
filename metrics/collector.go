package metrics

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
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

var collectorRegistry = make(map[Collector]bool)

func RegisterCollector(collector Collector) {
	collectorRegistry[collector] = true
}

// ================================= Collector Source =================================
var MetricsChanged = errors.New("Metrics of this collector have changed")

type CollectorSource struct {
	CollectInterval time.Duration
	SinkInterval    time.Duration
	IgnoredMetrics  []*regexp.Regexp

	collectors []Collector
}

func (config *CollectorSource) Start(wg *sync.WaitGroup, sink MetricSink) error {
	wg.Add(1)
	go config.run(wg, sink)
	return nil
}

func (config *CollectorSource) run(wg *sync.WaitGroup, sink MetricSink) {
	defer wg.Done()
	for {
		var collectWg sync.WaitGroup
		config.collect(&collectWg, sink)
		collectWg.Wait()
	}
}

func (config *CollectorSource) collect(wg *sync.WaitGroup, sink MetricSink) {
	config.initCollectors()
	metrics := config.filteredMetrics()
	sort.Strings(metrics)
	header, values, collectors := config.constructSample(metrics)
	log.Printf("Locally collecting %v metrics through %v collectors\n", len(metrics), len(collectors))

	stopper := NewStopper(len(collectors) + 1)
	for _, collector := range collectors {
		wg.Add(1)
		go config.updateCollector(wg, collector, stopper)
	}
	wg.Add(1)
	go config.sinkMetrics(wg, header, values, sink, stopper)
}

func (config *CollectorSource) initCollectors() {
	config.collectors = make([]Collector, 0, len(collectorRegistry))
	for collector, _ := range collectorRegistry {
		if err := collector.Init(); err != nil {
			log.Printf("Failed to initialize data collector %T: %v\n", collector, err)
			continue
		}
		if err := collector.Update(); err != nil {
			log.Printf("Failed to update data collector %T: %v\n", collector, err)
			continue
		}
		config.collectors = append(config.collectors, collector)
	}
}

func (config *CollectorSource) allMetrics() []string {
	var all []string
	for _, collector := range config.collectors {
		metrics := collector.SupportedMetrics()
		for _, metric := range metrics {
			all = append(all, metric)
		}
	}
	return all
}

func (config *CollectorSource) filteredMetrics() (filtered []string) {
	all := config.allMetrics()
	filtered = make([]string, 0, len(all))
	for _, metric := range all {
		matches := false
		for _, regex := range config.IgnoredMetrics {
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

func (config *CollectorSource) collectorFor(metric string) Collector {
	for _, collector := range config.collectors {
		if collector.SupportsMetric(metric) {
			return collector
		}
	}
	return nil
}

func (config *CollectorSource) constructSample(metrics []string) (Header, []Value, []Collector) {
	set := make(map[Collector]bool)

	header := make(Header, len(metrics))
	values := make([]Value, len(metrics))
	for i, metricName := range metrics {
		collector := config.collectorFor(metricName)
		if collector == nil {
			log.Println("No collector found for", metricName)
			continue
		}
		header[i] = metricName
		metric := Metric{metricName, i, values}

		if err := collector.Collect(&metric); err != nil {
			log.Printf("Error starting collector for %v: %v\n", metricName, err)
			continue
		}
		set[collector] = true
	}

	collectors := make([]Collector, 0, len(set))
	for col, _ := range set {
		collectors = append(collectors, col)
	}
	return header, values, collectors
}

func (config *CollectorSource) updateCollector(wg *sync.WaitGroup, collector Collector, stopper *Stopper) {
	defer wg.Done()
	for {
		err := collector.Update()
		if err == MetricsChanged {
			log.Printf("Metrics of %v have changed! Restarting metric collection.\n", collector)
			stopper.Stop()
			return
		} else if err != nil {
			log.Println("Warning: Collector update failed:", err)
		}
		if stopper.Stopped(config.CollectInterval) {
			return
		}
	}
}

func (config *CollectorSource) sinkMetrics(wg *sync.WaitGroup, header Header, values []Value, sink MetricSink, stopper *Stopper) {
	defer wg.Done()
	for {
		if err := sink.Header(header); err != nil {
			log.Printf("Warning: Failed to sink header for %v metrics: %v\n", len(header), err)
		} else {
			if stopper.IsStopped() {
				return
			}
			for {
				sample := Sample{
					time.Now(),
					values,
				}
				if err := sink.Sample(sample); err != nil {
					// When a sample fails, try sending the header again
					log.Printf("Warning: Failed to sink %v metrics: %v\n", len(values), err)
					break
				}
				if stopper.Stopped(config.SinkInterval) {
					return
				}
			}
		}
		if stopper.Stopped(config.SinkInterval) {
			return
		}
	}
}

// ================================= Abstract Collector =================================
type AbstractCollector struct {
	metrics []*CollectedMetric
	readers map[string]MetricReader // Must be filled in Init() implementations
	name    string
}

type CollectedMetric struct {
	*Metric
	MetricReader
}

type MetricReader func() Value

func (col *AbstractCollector) Reset(parent interface{}) {
	col.metrics = nil
	col.readers = nil
	col.name = fmt.Sprintf("%T", parent)
}

func (col *AbstractCollector) SupportedMetrics() (res []string) {
	res = make([]string, 0, len(col.readers))
	for metric, _ := range col.readers {
		res = append(res, metric)
	}
	return
}

func (col *AbstractCollector) SupportsMetric(metric string) bool {
	_, ok := col.readers[metric]
	return ok
}

func (col *AbstractCollector) Collect(metric *Metric) error {
	tags := make([]string, 0, len(col.readers))
	for metricName, reader := range col.readers {
		if metric.Name == metricName {
			col.metrics = append(col.metrics, &CollectedMetric{
				Metric:       metric,
				MetricReader: reader,
			})
			return nil
		}
		tags = append(tags, metric.Name)
	}
	return fmt.Errorf("Cannot handle metric %v, expected one of %v", metric.Name, tags)
}

func (col *AbstractCollector) UpdateMetrics() {
	for _, metric := range col.metrics {
		metric.Set(metric.MetricReader())
	}
}

func (col *AbstractCollector) String() string {
	return fmt.Sprintf("%s (%v metrics)", col.name, len(col.metrics))
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
