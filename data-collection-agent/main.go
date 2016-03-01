package main

import (
	"time"

	"github.com/citlab/monitoring/metrics"
)

func main() {
	met := metrics.AllMetrics()
	collectors := metrics.CollectMetrics(met...)
	for _, collector := range collectors {
		go metrics.UpdateCollector(collector, 100*time.Millisecond)
	}

	go metrics.SinkMetrics(met, &metrics.ConsoleSink{}, 500*time.Millisecond)
	metrics.SinkMetrics(met, &metrics.TcpSink{Endpoint: "127.0.0.1:9999"}, 500*time.Millisecond)
}
