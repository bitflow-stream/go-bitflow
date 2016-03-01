package metrics

import (
	"log"
	"strings"
	"time"
)

type Collector interface {
	Collect(metric *Metric) error
	Update() error
	SupportedMetrics() []string
}

var collectors map[string]Collector

func init() {
	collectors = make(map[string]Collector)
}

func UpdateCollector(collector Collector, interval time.Duration) {
	for {
		err := collector.Update()
		if err != nil {
			log.Println("Warning: Collector update failed:", err)
		}
		time.Sleep(interval)
	}
}

func registerCollector(prefix string, collector Collector) {
	if _, ok := collectors[prefix]; ok {
		log.Println("Warning: overriding registered collector for", prefix)
	}
	collectors[prefix] = collector
}

func CollectorFor(metric *Metric) Collector {
	for prefix, collector := range collectors {
		if strings.HasPrefix(metric.Tag, prefix) {
			return collector
		}
	}
	return nil
}

func CollectMetrics(metrics ...*Metric) []Collector {
	set := make(map[Collector]bool)

	for _, metric := range metrics {
		collector := CollectorFor(metric)
		if collector == nil {
			log.Println("Warning: no collector found for", metric.Tag)
			continue
		}
		collector.Collect(metric)
		set[collector] = true
	}

	result := make([]Collector, 0, len(set))
	for col, _ := range set {
		result = append(result, col)
	}
	return result
}

func AllMetrics() []*Metric {
	var all []*Metric
	for _, collector := range collectors {
		metrics := collector.SupportedMetrics()
		for _, metric := range metrics {
			all = append(all, &Metric{
				Tag: metric,
			})
		}
	}
	return all
}
