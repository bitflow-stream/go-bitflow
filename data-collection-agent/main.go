package main

import (
	"flag"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/data2go/sample-collector"
	"github.com/antongulenko/golib"
)

var (
	collect_local_interval = 500 * time.Millisecond
	sink_interval          = 500 * time.Millisecond
	active_retry_interval  = 1000 * time.Millisecond

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

	all_metrics          = false
	user_include_metrics golib.StringSlice
	user_exclude_metrics golib.StringSlice

	proc_collectors      golib.StringSlice
	proc_collector_regex golib.StringSlice
	proc_show_errors     = false
	proc_update_pids     = 1500 * time.Millisecond

	print_metrics = false
	libvirt_uri   = collector.LibvirtLocal() // collector.LibvirtSsh("host", "keyfile")
	ovsdb_host    = ""
)

var (
	includeMetricsRegexes []*regexp.Regexp
	excludeMetricsRegexes = []*regexp.Regexp{
		regexp.MustCompile("^net-proto/(UdpLite|IcmpMsg)"),                         // Some extended protocol-metrics
		regexp.MustCompile("^disk-io/...[0-9]"),                                    // Disk IO for specific partitions
		regexp.MustCompile("^disk-usage//.+/(used|free)$"),                         // All partitions except root
		regexp.MustCompile("^net-proto/tcp/(MaxConn|RtpAlgorithm|RtpMin|RtoMax)$"), // Some irrelevant TCP/IP settings
		regexp.MustCompile("^net-proto/ip/(DefaultTTL|Forwarding)$"),
	}
	marshallers = map[string]sample.MetricMarshaller{
		"":  new(sample.CsvMarshaller), // The default
		"c": new(sample.CsvMarshaller),
		"b": new(sample.BinaryMarshaller),
		"t": new(sample.TextMarshaller),
	}
)

func marshaller(format string) sample.MetricMarshaller {
	if marshaller, ok := marshallers[format]; !ok {
		log.Fatalf("Illegal data fromat %v, must be one of %v\n", format, supportedFormats)
		return nil
	} else {
		return marshaller
	}
}

func main() {
	flag.StringVar(&libvirt_uri, "libvirt", libvirt_uri, "Libvirt connection uri (default is local system)")
	flag.StringVar(&ovsdb_host, "ovsdb", ovsdb_host, "OVSDB host to connect to. Empty for localhost. Port is "+strconv.Itoa(collector.DefaultOvsdbPort))
	flag.BoolVar(&print_metrics, "metrics", print_metrics, "Print all available metrics and exit")
	flag.BoolVar(&all_metrics, "a", all_metrics, "Disable built-in filters on available metrics")
	flag.Var(&user_exclude_metrics, "exclude", "Metrics to exclude (only with -c, substring match)")
	flag.Var(&user_include_metrics, "include", "Metrics to include exclusively (only with -c, substring match)")

	flag.Var(&proc_collectors, "proc", "Processes to collect metrics for (substring match on entire command line)")
	flag.Var(&proc_collector_regex, "proc_regex", "Processes to collect metrics for (regex match on entire command line)")
	flag.BoolVar(&proc_show_errors, "proc_err", proc_show_errors, "Verbose: show errors encountered while getting process metrics")
	flag.DurationVar(&proc_update_pids, "proc_interval", proc_update_pids, "Interval for updating list of observed pids")

	flag.StringVar(&format_input, "i", format_input, "Data source format (does not apply to -c), one of "+supportedFormats)
	flag.BoolVar(&collect_local, "c", collect_local, "Data source: collect local samples")
	flag.DurationVar(&collect_local_interval, "ci", collect_local_interval, "Interval for collecting local samples")
	flag.DurationVar(&sink_interval, "si", sink_interval, "Interval for sinking (sending/printing/...) data when collecting local samples")
	flag.BoolVar(&collect_console, "C", collect_console, "Data source: read from stdin")
	flag.StringVar(&collect_file, "F", collect_file, "Data source: read from file")
	flag.StringVar(&collect_listen, "L", collect_listen, "Data source: receive samples by accepting a TCP connection")
	flag.StringVar(&collect_download, "D", collect_download, "Data source: receive samples by connecting to remote endpoint")

	flag.BoolVar(&sink_console, "p", sink_console, "Data sink: print to stdout")
	flag.StringVar(&format_console, "pf", format_console, "Data format for console output, one of "+supportedFormats)
	flag.StringVar(&sink_file, "f", sink_file, "Data sink: write data to file")
	flag.StringVar(&format_file, "ff", format_file, "Data format for file output, one of "+supportedFormats)
	flag.StringVar(&sink_connect, "s", sink_connect, "Data sink: send data to specified TCP endpoint")
	flag.StringVar(&format_connect, "sf", format_connect, "Data format for TCP output, one of "+supportedFormats)
	flag.StringVar(&sink_listen, "l", sink_listen, "Data sink: accept TCP connections for sending out data")
	flag.StringVar(&format_listen, "lf", format_listen, "Data format for TCP server output, one of "+supportedFormats)
	flag.Parse()

	defer golib.ProfileCpu()()

	// ====== Configure collectors
	collector.RegisterPsutilCollectors()
	collector.RegisterLibvirtCollector(libvirt_uri)
	collector.RegisterOvsdbCollector(ovsdb_host)
	if len(proc_collectors) > 0 || len(proc_collector_regex) > 0 {
		procRegex := make([]*regexp.Regexp, 0, len(proc_collectors))
		for _, substr := range proc_collectors {
			regex := regexp.MustCompile(regexp.QuoteMeta(substr))
			procRegex = append(procRegex, regex)
		}
		for _, regexStr := range proc_collector_regex {
			regex, err := regexp.Compile(regexStr)
			golib.Checkerr(err)
			procRegex = append(procRegex, regex)
		}
		collector.RegisterCollector(&collector.PsutilProcessCollector{
			CmdlineFilter:     procRegex,
			GroupName:         "vnf",
			PrintErrors:       proc_show_errors,
			PidUpdateInterval: proc_update_pids,
		})
	}

	if all_metrics {
		excludeMetricsRegexes = nil
	}
	for _, exclude := range user_exclude_metrics {
		excludeMetricsRegexes = append(excludeMetricsRegexes,
			regexp.MustCompile(regexp.QuoteMeta(exclude)))
	}
	for _, include := range user_include_metrics {
		includeMetricsRegexes = append(includeMetricsRegexes,
			regexp.MustCompile(regexp.QuoteMeta(include)))
	}

	// ====== Data format
	marshaller_console := marshaller(format_console)
	marshaller_connect := marshaller(format_connect)
	marshaller_listen := marshaller(format_listen)
	marshaller_file := marshaller(format_file)
	unmarshaller := marshaller(format_input)

	// ====== Initialize source(s)
	var source sample.MetricSource
	setSource := func(set bool, theSource sample.MetricSource) {
		if set {
			if source != nil {
				log.Fatalln("Please provide only one data source")
			}
			source = theSource
		}
	}
	collector := &collector.CollectorSource{
		CollectInterval: collect_local_interval,
		SinkInterval:    sink_interval,
		ExcludeMetrics:  excludeMetricsRegexes,
		IncludeMetrics:  includeMetricsRegexes,
	}
	if print_metrics {
		collector.PrintMetrics()
		return
	}
	setSource(collect_local, collector)
	setSource(collect_console, new(sample.ConsoleSource))
	setSource(collect_listen != "", sample.NewTcpListenerSource(collect_listen))
	setSource(collect_download != "", &sample.TCPSource{
		RemoteAddr:    collect_download,
		RetryInterval: active_retry_interval,
	})
	setSource(collect_file != "", &sample.FileSource{
		FileTransport: sample.FileTransport{Filename: collect_file}})
	if source == nil {
		log.Println("No data source provided, no data will be generated.")
	}

	// ====== Initialize sink(s) and tasks
	var sinks sample.AggregateSink
	var marshallers []sample.Marshaller
	if sink_console {
		sinks = append(sinks, new(sample.ConsoleSink))
		marshallers = append(marshallers, marshaller_console)
	}
	if sink_connect != "" {
		sinks = append(sinks, &sample.TCPSink{Endpoint: sink_connect})
		marshallers = append(marshallers, marshaller_connect)
	}
	if sink_listen != "" {
		sinks = append(sinks, sample.NewTcpListenerSink(sink_listen))
		marshallers = append(marshallers, marshaller_listen)
	}
	if sink_file != "" {
		sinks = append(sinks, &sample.FileSink{FileTransport: sample.FileTransport{Filename: sink_file}})
		marshallers = append(marshallers, marshaller_file)
	}
	if len(sinks) == 0 {
		log.Println("No data sinks selected, data will not be output anywhere.")
	}

	// ====== Start and wait
	tasks := golib.NewTaskGroup(source)
	for i, sink := range sinks {
		sink.SetMarshaller(marshallers[i])
		tasks.Add(sink)
	}
	if source != nil {
		source.SetSink(sinks)
		if unmarshallingSource, ok := source.(sample.UnmarshallingMetricSource); ok {
			unmarshallingSource.SetUnmarshaller(unmarshaller)
		}
	}
	log.Println("Press Ctrl-C to interrupt")
	tasks.Add(&golib.NoopTask{golib.ExternalInterrupt(), "external interrupt"})
	tasks.WaitAndExit()
}
