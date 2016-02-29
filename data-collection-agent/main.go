package main

import (
	"fmt"
	"log"
	"time"

	"github.com/citlab/monitoring/metrics"
)

const (
	metric_config_url = "https://raw.githubusercontent.com/citlab/monitoring/master/metrics.ini"
	metric_local_file = "/home/anton/.gvm/pkgsets/go1.4/anton/src/github.com/citlab/monitoring/metrics.ini"
)

func LoadMetricConfig() {
	//	err := metrics.DownloadTagConfig(metric_config_url)
	err := metrics.ReadTagConfig(metric_local_file)
	if err != nil {
		log.Fatalln("Reading tag config failed:", err)
	}
	fmt.Println("Monitoring config:", metrics.Tags)
}

func main() {
	LoadMetricConfig()

	met := []*metrics.Metric{
		{Tag: metrics.Tags["mem/free"]},
		{Tag: metrics.Tags["mem/used"]},
		{Tag: metrics.Tags["mem/percent"]},
		{Tag: metrics.Tags["cpu"]},
	}
	collectors := metrics.CollectMetrics(met...)
	for _, collector := range collectors {
		go metrics.UpdateCollector(collector, time.Second)
	}

	//	metrics.SinkMetrics(met, &metrics.ConsoleSink{}, 500*time.Millisecond)
	metrics.SinkMetrics(met, &metrics.TcpSink{Endpoint: "127.0.0.1:9999"}, 500*time.Millisecond)
}
