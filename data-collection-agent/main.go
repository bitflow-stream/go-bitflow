package main

import (
	"log"
	"time"

	"github.com/citlab/monitoring/metrics"
)

var (
	ignoredMetrics = []string{
	//		"net-proto",
	}
)

func main() {
	if err := metrics.InitCollectors(); err != nil {
		log.Fatalln(err)
	}
	met := metrics.AllMetrics()
	met = metrics.FilterMetrics(met, ignoredMetrics)

	collectors, err := metrics.CollectMetrics(met...)
	if err != nil {
		log.Fatalln(err)
	}

	for _, collector := range collectors {
		go metrics.UpdateCollector(collector, 100*time.Millisecond)
	}

	metrics.SinkMetrics(met, &metrics.ConsoleSink{}, 500*time.Millisecond)
	//	metrics.SinkMetrics(met, &metrics.TcpSink{Endpoint: "127.0.0.1:9999"}, 500*time.Millisecond)
}
