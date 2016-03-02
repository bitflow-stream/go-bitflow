package metrics

import (
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"
)

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

func SinkMetrics(wg *sync.WaitGroup, metrics []*Metric, target MetricSink, interval time.Duration) {
	defer wg.Done()
	for {
		if err := target.CycleStart(); err != nil {
			log.Printf("Warning: Failed to sink %v metrics: %v\n", len(metrics), err)
		} else {
			for _, metric := range metrics {
				err := target.Sink(metric)
				if err != nil {
					log.Printf("Warning: Failed to sink metric %v: %v\n", metric, err)
					break
				}
			}
			target.CycleFinish()
		}
		time.Sleep(interval)
	}
}
