package main

import "github.com/citlab/monitoring/metrics"

const (
	listen = "127.0.0.1:9999"
)

func main() {
	sink := &metrics.ConsoleSink{}
	metrics.ReceiveMetrics(listen, sink)
}
