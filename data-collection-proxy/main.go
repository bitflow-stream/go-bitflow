package main

import (
	agent "github.com/citlab/monitoring/data-collection-agent"
	"github.com/citlab/monitoring/metrics"
)

const (
	listen = "127.0.0.1:9999"
)

func main() {
	agent.LoadMetricConfig()
	sink := &metrics.ConsoleSink{}
	metrics.ReceiveMetrics(listen, sink)
}
