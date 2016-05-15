package collector

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

// ==================== Metric ====================
type Metric struct {
	Name   string
	index  int
	sample []sample.Value
}

func (metric *Metric) Set(val sample.Value) {
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

const FailedCollectorCheckInterval = 5 * time.Second
const FilteredCollectorCheckInterval = 30 * time.Second

type CollectorSource struct {
	CollectInterval time.Duration
	SinkInterval    time.Duration
	ExcludeMetrics  []*regexp.Regexp
	IncludeMetrics  []*regexp.Regexp

	collectors         []Collector
	failedCollectors   []Collector
	filteredCollectors []Collector

	sink     sample.MetricSink
	loopTask golib.Task
}

func (source *CollectorSource) String() string {
	return "CollectorSource"
}

func (source *CollectorSource) SetSink(sink sample.MetricSink) {
	source.sink = sink
}

func (source *CollectorSource) Start(wg *sync.WaitGroup) golib.StopChan {
	// TODO integrate golib.StopChan/LoopTask and golib.Stopper
	source.loopTask = golib.NewLoopTask("CollectorSource", func(stop golib.StopChan) {
		var collectWg sync.WaitGroup
		stopper := source.collect(&collectWg)
		select {
		case <-stopper.Wait():
		case <-stop:
		}
		stopper.Stop()
		collectWg.Wait()
	})
	return source.loopTask.Start(wg)
}

func (source *CollectorSource) Stop() {
	source.loopTask.Stop()
}

func (source *CollectorSource) collect(wg *sync.WaitGroup) *golib.Stopper {
	source.initCollectors()
	metrics := source.FilteredMetrics()
	sort.Strings(metrics)
	header, values, collectors := source.constructSample(metrics)
	log.Printf("Locally collecting %v metrics through %v collectors\n", len(metrics), len(collectors))

	stopper := golib.NewStopper()
	for _, collector := range source.collectors {
		var interval time.Duration
		if _, ok := collectors[collector]; ok {
			interval = source.CollectInterval
		} else {
			interval = FilteredCollectorCheckInterval
		}
		wg.Add(1)
		go source.updateCollector(wg, collector, stopper, interval)
	}
	for _, failed := range source.failedCollectors {
		wg.Add(1)
		go source.watchFailedCollector(wg, failed, stopper)
	}
	wg.Add(1)
	go source.sinkMetrics(wg, header, values, source.sink, stopper)
	return stopper
}

func (source *CollectorSource) initCollectors() {
	source.collectors = make([]Collector, 0, len(collectorRegistry))
	source.failedCollectors = make([]Collector, 0, len(collectorRegistry))
	source.filteredCollectors = make([]Collector, 0, len(collectorRegistry))
	for collector, _ := range collectorRegistry {
		if err := source.initCollector(collector); err != nil {
			log.Printf("Failed to initialize data collector %v: %v\n", collector, err)
			source.failedCollectors = append(source.failedCollectors, collector)
			continue
		}
		source.collectors = append(source.collectors, collector)
	}
}

func (source *CollectorSource) initCollector(collector Collector) error {
	if err := collector.Init(); err != nil {
		return err
	}
	if err := collector.Update(); err != nil {
		return err
	}
	return nil
}

func (source *CollectorSource) AllMetrics() []string {
	var all []string
	for _, collector := range source.collectors {
		metrics := collector.SupportedMetrics()
		for _, metric := range metrics {
			all = append(all, metric)
		}
	}
	return all
}

func (source *CollectorSource) FilteredMetrics() (filtered []string) {
	all := source.AllMetrics()
	filtered = make([]string, 0, len(all))
	for _, metric := range all {
		excluded := false
		for _, regex := range source.ExcludeMetrics {
			if excluded = regex.MatchString(metric); excluded {
				break
			}
		}
		if !excluded && len(source.IncludeMetrics) > 0 {
			excluded = true
			for _, regex := range source.IncludeMetrics {
				if excluded = !regex.MatchString(metric); !excluded {
					break
				}
			}
		}
		if !excluded {
			filtered = append(filtered, metric)
		}
	}
	return
}

func (source *CollectorSource) collectorFor(metric string) Collector {
	for _, collector := range source.collectors {
		if collector.SupportsMetric(metric) {
			return collector
		}
	}
	return nil
}

func (source *CollectorSource) constructSample(metrics []string) (sample.Header, []sample.Value, map[Collector]bool) {
	set := make(map[Collector]bool)

	header := make(sample.Header, len(metrics))
	values := make([]sample.Value, len(metrics))
	for i, metricName := range metrics {
		collector := source.collectorFor(metricName)
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
	return header, values, set
}

func (source *CollectorSource) updateCollector(wg *sync.WaitGroup, collector Collector, stopper *golib.Stopper, interval time.Duration) {
	defer wg.Done()
	for {
		err := collector.Update()
		if err == MetricsChanged {
			log.Printf("Metrics of %v have changed! Restarting metric collection.\n", collector)
			stopper.Stop()
			return
		} else if err != nil {
			log.Println("Warning: Update of", collector, "failed:", err)
		}
		if stopper.Stopped(interval) {
			return
		}
	}
}

func (source *CollectorSource) watchFailedCollector(wg *sync.WaitGroup, collector Collector, stopper *golib.Stopper) {
	defer wg.Done()
	for {
		if stopper.Stopped(FailedCollectorCheckInterval) {
			return
		}
		if err := source.initCollector(collector); err == nil {
			log.Println("Collector", collector, "is not failing anymore. Restarting metric collection.")
			stopper.Stop()
			return
		}
	}
}

func (source *CollectorSource) sinkMetrics(wg *sync.WaitGroup, header sample.Header, values []sample.Value, sink sample.MetricSink, stopper *golib.Stopper) {
	defer wg.Done()
	for {
		if err := sink.Header(header); err != nil {
			log.Printf("Warning: Failed to sink header for %v metrics: %v\n", len(header), err)
		} else {
			if stopper.IsStopped() {
				return
			}
			for {
				sample := sample.Sample{
					time.Now(),
					values,
				}
				if err := sink.Sample(sample); err != nil {
					// When a sample fails, try sending the header again
					log.Printf("Warning: Failed to sink %v metrics: %v\n", len(values), err)
					break
				}
				if stopper.Stopped(source.SinkInterval) {
					return
				}
			}
		}
		if stopper.Stopped(source.SinkInterval) {
			return
		}
	}
}

func (source *CollectorSource) PrintMetrics() {
	source.initCollectors()
	all := source.AllMetrics()
	filtered := source.FilteredMetrics()
	sort.Strings(all)
	sort.Strings(filtered)
	i := 0
	for _, metric := range all {
		isIncluded := i < len(filtered) && filtered[i] == metric
		if isIncluded {
			i++
		}
		if !isIncluded {
			fmt.Println(metric, "(excluded)")
		} else {
			fmt.Println(metric)
		}
	}
}

// ================================= Abstract Collector =================================
type AbstractCollector struct {
	metrics []*CollectedMetric
	readers map[string]MetricReader // Must be filled in Init() implementations
	notify  map[string]CollectNotification
	name    string
}

type CollectedMetric struct {
	*Metric
	MetricReader
}

type CollectNotification func()
type MetricReader func() sample.Value

func (source *AbstractCollector) Reset(parent interface{}) {
	source.metrics = nil
	source.readers = nil
	source.notify = make(map[string]CollectNotification)
	source.name = fmt.Sprintf("%T", parent)
}

func (source *AbstractCollector) SupportedMetrics() (res []string) {
	res = make([]string, 0, len(source.readers))
	for metric, _ := range source.readers {
		res = append(res, metric)
	}
	return
}

func (source *AbstractCollector) SupportsMetric(metric string) bool {
	_, ok := source.readers[metric]
	return ok
}

func (source *AbstractCollector) Collect(metric *Metric) error {
	tags := make([]string, 0, len(source.readers))
	for metricName, reader := range source.readers {
		if metric.Name == metricName {
			source.metrics = append(source.metrics, &CollectedMetric{
				Metric:       metric,
				MetricReader: reader,
			})
			if notifier, ok := source.notify[metric.Name]; ok {
				notifier()
			}
			return nil
		}
		tags = append(tags, metric.Name)
	}
	return fmt.Errorf("Cannot handle metric %v, expected one of %v", metric.Name, tags)
}

func (source *AbstractCollector) UpdateMetrics() {
	for _, metric := range source.metrics {
		metric.Set(metric.MetricReader())
	}
}

func (source *AbstractCollector) String() string {
	return fmt.Sprintf("%s (%v metrics)", source.name, len(source.metrics))
}

// ================================= Goroutine pool for collector tasks =================================
type CollectorTask func() error

type CollectorTaskPolicyType int

var CollectorTaskPolicy = CollectorTasksUntilError

const (
	CollectorTasksSequential = CollectorTaskPolicyType(0)
	CollectorTasksParallel   = CollectorTaskPolicyType(1)
	CollectorTasksUntilError = CollectorTaskPolicyType(2)
)

type CollectorTasks []CollectorTask

func (pool CollectorTasks) Run() error {
	switch CollectorTaskPolicy {
	case CollectorTasksSequential:
		return pool.RunSequential()
	case CollectorTasksParallel:
		return pool.RunParallel()
	default:
		fallthrough
	case CollectorTasksUntilError:
		return pool.RunUntilError()
	}
}

func (pool CollectorTasks) RunParallel() error {
	var wg sync.WaitGroup
	var errors golib.MultiError
	var errorsLock sync.Mutex
	wg.Add(len(pool))
	for _, task := range pool {
		go func(task CollectorTask) {
			defer wg.Done()
			err := task()
			errorsLock.Lock()
			defer errorsLock.Unlock()
			errors.Add(err)
		}(task)
	}
	wg.Wait()
	return errors.NilOrError()
}

func (pool CollectorTasks) RunSequential() error {
	var errors golib.MultiError
	for _, task := range pool {
		err := task()
		errors.Add(err)
	}
	return errors.NilOrError()
}

func (pool CollectorTasks) RunUntilError() error {
	for _, task := range pool {
		if err := task(); err != nil {
			return err
		}
	}
	return nil
}
