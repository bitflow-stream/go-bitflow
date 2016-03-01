package main

import (
	"log"
	"time"

	"github.com/citlab/monitoring/metrics"
)

const (
	collect_interval = 100 * time.Millisecond
	print_interval   = 500 * time.Millisecond
	send_interval    = 500 * time.Millisecond
	send_endpoint    = "127.0.0.1:9999"
	do_print         = true
	do_send          = false
)

var (
	ignoredMetrics = []string{
	//		"net-proto",
	}
)

func main() {
	// ====== Initialize
	if err := metrics.InitCollectors(); err != nil {
		log.Fatalln(err)
	}
	met := metrics.AllMetrics()
	met = metrics.FilterMetrics(met, ignoredMetrics)

	// ====== Collect and update
	collectors, err := metrics.CollectMetrics(met...)
	if err != nil {
		log.Fatalln(err)
	}
	for _, collector := range collectors {
		go metrics.UpdateCollector(collector, collect_interval)
	}

	// ====== Print/Send results
	console := func() {
		metrics.SinkMetrics(met, &metrics.ConsoleSink{}, print_interval)
	}
	send := func() {
		metrics.SinkMetrics(met, &metrics.TcpSink{Endpoint: send_endpoint}, send_interval)
	}
	switch {
	case do_print && do_send:
		go console()
		send()
	case do_print:
		console()
	case do_send:
		send()
	default:
		log.Fatalln("Need at least one data sink (console or tcp)")
	}
}
