package main

import (
	"flag"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/citlab/monitoring/metrics"
)

var (
	collect_local_interval = 300 * time.Millisecond
	sink_interval          = 500 * time.Millisecond

	collect_local    = false
	collect_listen   = ""
	collect_download = ""

	sink_console = false
	sink_connect = ""
	sink_listen  = ""

	active_retry_interval = 1000 * time.Millisecond
)

var (
	ignoredMetrics = []*regexp.Regexp{
		regexp.MustCompile("^net-proto/(UdpLite|IcmpMsg)"), // Some extended protocol-metrics
		regexp.MustCompile("^disk-io/...[0-9]"),            // Disk IO for specific partitions
		regexp.MustCompile("^disk-usage//.+/(used|free)$"), // All partitions except root
	}
)

func main() {
	flag.BoolVar(&collect_local, "c", collect_local, "Data source: collect local metrics")
	flag.DurationVar(&collect_local_interval, "ci", collect_local_interval, "Interval for collecting local metrics")
	flag.DurationVar(&sink_interval, "si", sink_interval, "Interval for sinking (sending/printing/...) data when collecting local metrics")
	flag.StringVar(&collect_listen, "L", collect_listen, "Data source: receive metrics by accepting TCP connections")
	flag.StringVar(&collect_download, "d", collect_download, "Data source: receive metrics by connecting to remote endpoint")
	flag.BoolVar(&sink_console, "p", sink_console, "Data sink: print to console")
	flag.StringVar(&sink_connect, "s", sink_connect, "Data sink: send data to specified endpoint")
	flag.StringVar(&sink_listen, "l", sink_listen, "Data sink: accept connections to send data on specified endpoint")
	flag.Parse()

	// ====== Initialize sink(s)
	var sinks metrics.AggregateSink
	if sink_console {
		sinks = append(sinks, new(metrics.ConsoleSink))
	}
	if sink_connect != "" {
		sinks = append(sinks, &metrics.ActiveTcpSink{Endpoint: sink_connect})
	}
	if sink_listen != "" {
		sinks = append(sinks, &metrics.PassiveTcpSink{Endpoint: sink_listen})
	}
	if len(sinks) == 0 {
		log.Println("No data sinks selected, data will not be output anywhere.")
	}

	// ====== Initialize source(s)
	var sources metrics.AggregateSource
	if collect_local {
		sources = append(sources, new(CollectorSource))
	}
	if collect_listen != "" {
		sources = append(sources, &metrics.ReceiveMetricSource{
			ListenEndpoint: collect_listen,
		})
	}
	if collect_download != "" {
		sources = append(sources, &metrics.DownloadMetricSource{
			RemoteAddr:    collect_download,
			RetryInterval: active_retry_interval,
		})
	}
	if len(sources) == 0 {
		log.Println("No data sources selected, no data will be generated.")
	}

	// ====== Start and wait
	var wg sync.WaitGroup
	if err := sinks.Start(&wg); err != nil {
		log.Fatalln(err)
	}
	if err := sources.Start(&wg, sinks); err != nil {
		log.Fatalln(err)
	}
	wg.Wait()
}

type CollectorSource struct {
}

func (source *CollectorSource) Start(wg *sync.WaitGroup, sink metrics.MetricSink) error {
	if err := metrics.InitCollectors(); err != nil {
		return err
	}
	met := metrics.AllMetrics()
	met = metrics.FilterMetrics(met, ignoredMetrics)
	log.Printf("Locally collecting %v metrics\n", len(met))

	collectors, err := metrics.CollectMetrics(met...)
	if err != nil {
		return err
	}
	for _, collector := range collectors {
		wg.Add(1)
		go metrics.UpdateCollector(wg, collector, collect_local_interval)
	}

	wg.Add(1)
	go metrics.SinkMetrics(wg, met, sink, sink_interval)
	return nil
}
