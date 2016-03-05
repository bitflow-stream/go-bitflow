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
	collect_console  = false
	collect_listen   = ""
	collect_download = ""
	collect_file     = ""

	sink_console = false
	sink_connect = ""
	sink_listen  = ""
	sink_file    = ""

	format_input  = ""
	format_output = ""
	format_both   = ""

	active_retry_interval = 1000 * time.Millisecond
)

var (
	ignoredMetrics = []*regexp.Regexp{
		regexp.MustCompile("^net-proto/(UdpLite|IcmpMsg)"), // Some extended protocol-metrics
		regexp.MustCompile("^disk-io/...[0-9]"),            // Disk IO for specific partitions
		regexp.MustCompile("^disk-usage//.+/(used|free)$"), // All partitions except root
	}
	marshallers = map[string]metrics.MetricMarshaller{
		"":  new(metrics.CsvMarshaller), // The default
		"c": new(metrics.CsvMarshaller),
		"b": new(metrics.BinaryMarshaller),
		"t": new(metrics.TextMarshaller),
	}
)

func main() {
	supportedFormats := "(c=CSV, b=binary, t=text)"
	flag.StringVar(&format_input, "i", "", "Data source format, one of "+supportedFormats)
	flag.StringVar(&format_output, "o", "", "Data sink format, one of "+supportedFormats)
	flag.StringVar(&format_both, "io", "", "Shortcut for -i and -o")

	flag.BoolVar(&collect_local, "c", collect_local, "Data source: collect local metrics")
	flag.DurationVar(&collect_local_interval, "ci", collect_local_interval, "Interval for collecting local metrics")
	flag.DurationVar(&sink_interval, "si", sink_interval, "Interval for sinking (sending/printing/...) data when collecting local metrics")
	flag.BoolVar(&collect_console, "C", collect_console, "Data source: read from stdin")
	flag.StringVar(&collect_file, "F", collect_file, "Data source: read from file")
	flag.StringVar(&collect_listen, "L", collect_listen, "Data source: receive metrics by accepting a TCP connection")
	flag.StringVar(&collect_download, "d", collect_download, "Data source: receive metrics by connecting to remote endpoint")

	flag.BoolVar(&sink_console, "p", sink_console, "Data sink: print to stdout")
	flag.StringVar(&sink_file, "f", sink_file, "Data sink: write data to file")
	flag.StringVar(&sink_connect, "s", sink_connect, "Data sink: send data to specified endpoint")
	flag.StringVar(&sink_listen, "l", sink_listen, "Data sink: accept connections for sending out data")
	flag.Parse()

	// ====== Configure collectors
	metrics.RegisterLibvirtCollectors(metrics.LibvirtLocal())

	// ====== Data format
	if format_both != "" {
		if format_input != "" || format_output != "" {
			log.Fatalln("Please provide only one of [-i/-o] or [-io]")
		}
		format_input = format_both
		format_output = format_both
	}
	var marshaller metrics.Marshaller
	var unmarshaller metrics.Unmarshaller
	var ok bool
	if marshaller, ok = marshallers[format_output]; !ok {
		log.Fatalf("Illegal output fromat %v, must be one of %v\n", format_output, supportedFormats)
	}
	if unmarshaller, ok = marshallers[format_input]; !ok {
		log.Fatalf("Illegal input fromat %v, must be one of %v\n", format_output, supportedFormats)
	}
	log.Printf("Reading %v data, writing %v data\n", unmarshaller, marshaller)

	// ====== Initialize sink(s)
	var sinks metrics.AggregateSink
	if sink_console {
		sinks = append(sinks, new(metrics.ConsoleSink))
	}
	if sink_connect != "" {
		sinks = append(sinks, &metrics.TCPSink{Endpoint: sink_connect})
	}
	if sink_listen != "" {
		sinks = append(sinks, &metrics.TCPListenerSink{Endpoint: sink_listen})
	}
	if sink_file != "" {
		sink := &metrics.FileSink{FileTransport: metrics.FileTransport{Filename: sink_file}}
		sinks = append(sinks, sink)
		defer sink.Close() // Ignore error
	}
	if len(sinks) == 0 {
		log.Println("No data sinks selected, data will not be output anywhere.")
	}

	// ====== Initialize source(s)
	var source metrics.MetricSource
	setSource := func(set bool, theSource metrics.MetricSource) {
		if set {
			if source != nil {
				log.Fatalln("Please provide only one data source")
			}
			source = theSource
		}
	}
	setSource(collect_local, &metrics.CollectorSource{
		CollectInterval: collect_local_interval,
		SinkInterval:    sink_interval,
		IgnoredMetrics:  ignoredMetrics,
	})
	setSource(collect_console, new(metrics.ConsoleSource))
	setSource(collect_listen != "", &metrics.TCPListenerSource{
		ListenEndpoint: collect_listen,
	})
	setSource(collect_download != "", &metrics.TCPSource{
		RemoteAddr:    collect_download,
		RetryInterval: active_retry_interval,
	})
	if collect_file != "" {
		fileSource := &metrics.FileSource{FileTransport: metrics.FileTransport{Filename: collect_file}}
		setSource(true, fileSource)
		defer fileSource.Close() // Ignore error
	}
	if source == nil {
		log.Println("No data source provided, no data will be generated.")
	}

	// ====== Start and wait
	var wg sync.WaitGroup
	if err := sinks.Start(&wg, marshaller); err != nil {
		log.Fatalln(err)
	}
	if source != nil {
		if err := source.Start(&wg, unmarshaller, sinks); err != nil {
			log.Fatalln(err)
		}
	}
	wg.Wait()
}
