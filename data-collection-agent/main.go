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

	supportedFormats = "(c=CSV, b=binary, t=text)"
	format_input     = "b"
	format_console   = "t"
	format_connect   = "b"
	format_listen    = "b"
	format_file      = "c"

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

func marshaller(format string) metrics.MetricMarshaller {
	if marshaller, ok := marshallers[format]; !ok {
		log.Fatalf("Illegal data fromat %v, must be one of %v\n", format, supportedFormats)
		return nil
	} else {
		return marshaller
	}
}

func main() {
	flag.StringVar(&format_input, "i", format_input, "Data source format (does not apply to -c), one of "+supportedFormats)

	flag.BoolVar(&collect_local, "c", collect_local, "Data source: collect local metrics")
	flag.DurationVar(&collect_local_interval, "ci", collect_local_interval, "Interval for collecting local metrics")
	flag.DurationVar(&sink_interval, "si", sink_interval, "Interval for sinking (sending/printing/...) data when collecting local metrics")
	flag.BoolVar(&collect_console, "C", collect_console, "Data source: read from stdin")
	flag.StringVar(&collect_file, "F", collect_file, "Data source: read from file")
	flag.StringVar(&collect_listen, "L", collect_listen, "Data source: receive metrics by accepting a TCP connection")
	flag.StringVar(&collect_download, "D", collect_download, "Data source: receive metrics by connecting to remote endpoint")

	flag.BoolVar(&sink_console, "p", sink_console, "Data sink: print to stdout")
	flag.StringVar(&format_console, "pf", format_console, "Data format for console output, one of "+supportedFormats)
	flag.StringVar(&sink_file, "f", sink_file, "Data sink: write data to file")
	flag.StringVar(&format_file, "ff", format_file, "Data format for file output, one of "+supportedFormats)
	flag.StringVar(&sink_connect, "s", sink_connect, "Data sink: send data to specified TCP endpoint")
	flag.StringVar(&format_connect, "sf", format_connect, "Data format for TCP output, one of "+supportedFormats)
	flag.StringVar(&sink_listen, "l", sink_listen, "Data sink: accept TCP connections for sending out data")
	flag.StringVar(&format_listen, "lf", format_listen, "Data format for TCP server output, one of "+supportedFormats)
	flag.Parse()

	// ====== Configure collectors
	metrics.RegisterPsutilCollectors()
	metrics.RegisterLibvirtCollectors(metrics.LibvirtLocal())

	// ====== Data format
	marshaller_console := marshaller(format_console)
	marshaller_connect := marshaller(format_connect)
	marshaller_listen := marshaller(format_listen)
	marshaller_file := marshaller(format_file)
	unmarshaller := marshaller(format_input)

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

	// ====== Initialize sink(s)
	var sinks metrics.AggregateSink
	var marshallers []metrics.Marshaller
	if sink_console {
		sinks = append(sinks, new(metrics.ConsoleSink))
		marshallers = append(marshallers, marshaller_console)
	}
	if sink_connect != "" {
		sinks = append(sinks, &metrics.TCPSink{Endpoint: sink_connect})
		marshallers = append(marshallers, marshaller_connect)
	}
	if sink_listen != "" {
		sinks = append(sinks, &metrics.TCPListenerSink{Endpoint: sink_listen})
		marshallers = append(marshallers, marshaller_listen)
	}
	if sink_file != "" {
		sink := &metrics.FileSink{FileTransport: metrics.FileTransport{Filename: sink_file}}
		defer sink.Close() // Ignore error
		sinks = append(sinks, sink)
		marshallers = append(marshallers, marshaller_file)
	}
	if len(sinks) == 0 {
		log.Println("No data sinks selected, data will not be output anywhere.")
	}

	// ====== Start and wait
	var wg sync.WaitGroup
	for i, sink := range sinks {
		if err := sink.Start(&wg, marshallers[i]); err != nil {
			log.Fatalln(err)
		}
	}
	if source != nil {
		if unmarshallingSource, ok := source.(metrics.UnmarshallingMetricSource); ok {
			unmarshallingSource.SetUnmarshaller(unmarshaller)
		}
		if err := source.Start(&wg, sinks); err != nil {
			log.Fatalln(err)
		}
	}
	wg.Wait()
}
