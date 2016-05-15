package pipeline

import (
	"flag"
	"log"
	"time"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

const (
	tcp_download_retry_interval = 1000 * time.Millisecond
	supported_formats           = "(c=CSV, b=binary, t=text)"
)

var (
	marshallers = map[string]sample.MetricMarshaller{
		"":  new(sample.CsvMarshaller), // The default
		"c": new(sample.CsvMarshaller),
		"b": new(sample.BinaryMarshaller),
		"t": new(sample.TextMarshaller),
	}
)

func marshaller(format string) sample.MetricMarshaller {
	if marshaller, ok := marshallers[format]; !ok {
		log.Fatalf("Illegal data fromat %v, must be one of %v\n", format, supported_formats)
		return nil
	} else {
		return marshaller
	}
}

type SamplePipeline struct {
	Sinks sample.AggregateSink
	Tasks *golib.TaskGroup

	source      sample.MetricSource
	sinks       sample.AggregateSink
	marshallers []sample.Marshaller

	read_console      bool
	read_tcp_listen   string
	read_tcp_download string
	read_file         string
	sink_console      bool
	sink_connect      string
	sink_listen       string
	sink_file         string
	format_input      string
	format_console    string
	format_connect    string
	format_listen     string
	format_file       string
}

func (p *SamplePipeline) ParseFlags() {
	flag.StringVar(&p.format_input, "i", "b", "Data source format, one of "+supported_formats)
	flag.BoolVar(&p.read_console, "C", false, "Data source: read from stdin")
	flag.StringVar(&p.read_file, "F", "", "Data source: read from file")
	flag.StringVar(&p.read_tcp_listen, "L", "", "Data source: receive samples by accepting a TCP connection")
	flag.StringVar(&p.read_tcp_download, "D", "", "Data source: receive samples by connecting to remote endpoint")

	flag.BoolVar(&p.sink_console, "p", false, "Data sink: print to stdout")
	flag.StringVar(&p.format_console, "pf", "t", "Data format for console output, one of "+supported_formats)
	flag.StringVar(&p.sink_file, "f", "", "Data sink: write data to file")
	flag.StringVar(&p.format_file, "ff", "c", "Data format for file output, one of "+supported_formats)
	flag.StringVar(&p.sink_connect, "s", "", "Data sink: send data to specified TCP endpoint")
	flag.StringVar(&p.format_connect, "sf", "b", "Data format for TCP output, one of "+supported_formats)
	flag.StringVar(&p.sink_listen, "l", "", "Data sink: accept TCP connections for sending out data")
	flag.StringVar(&p.format_listen, "lf", "b", "Data format for TCP server output, one of "+supported_formats)
}

func (p *SamplePipeline) Init() {
	// ====== Data format
	marshaller_console := marshaller(p.format_console)
	marshaller_connect := marshaller(p.format_connect)
	marshaller_listen := marshaller(p.format_listen)
	marshaller_file := marshaller(p.format_file)

	// ====== Initialize source(s)
	setSource := func(set bool, source sample.MetricSource) {
		if set {
			p.SetSource(source)
		}
	}
	setSource(p.read_console, new(sample.ConsoleSource))
	setSource(p.read_tcp_listen != "", sample.NewTcpListenerSource(p.read_tcp_listen))
	setSource(p.read_tcp_download != "", &sample.TCPSource{
		RemoteAddr:    p.read_tcp_download,
		RetryInterval: tcp_download_retry_interval,
	})
	setSource(p.read_file != "", &sample.FileSource{
		FileTransport: sample.FileTransport{Filename: p.read_file}})

	// ====== Initialize sink(s) and tasks
	if p.sink_console {
		p.sinks = append(p.sinks, new(sample.ConsoleSink))
		p.marshallers = append(p.marshallers, marshaller_console)
	}
	if p.sink_connect != "" {
		p.sinks = append(p.sinks, &sample.TCPSink{Endpoint: p.sink_connect})
		p.marshallers = append(p.marshallers, marshaller_connect)
	}
	if p.sink_listen != "" {
		p.sinks = append(p.sinks, sample.NewTcpListenerSink(p.sink_listen))
		p.marshallers = append(p.marshallers, marshaller_listen)
	}
	if p.sink_file != "" {
		p.sinks = append(p.sinks, &sample.FileSink{FileTransport: sample.FileTransport{Filename: p.sink_file}})
		p.marshallers = append(p.marshallers, marshaller_file)
	}

	// ====== Task group
	p.Tasks = golib.NewTaskGroup(p.source)
	for i, sink := range p.sinks {
		sink.SetMarshaller(p.marshallers[i])
		p.Tasks.Add(sink)
	}
}

func (p *SamplePipeline) SetSource(source sample.MetricSource) {
	if p.source != nil {
		log.Fatalln("Please provide only one data source")
	}
	p.source = source
}

func (p *SamplePipeline) StartAndWait() {
	if p.source == nil {
		log.Println("No data source provided, no data will be received.")
	}
	if len(p.sinks) == 0 {
		log.Println("No data sinks selected, data will not be output anywhere.")
	}

	if p.source != nil {
		p.source.SetSink(p.sinks)
		if unmarshallingSource, ok := p.source.(sample.UnmarshallingMetricSource); ok {
			unmarshallingSource.SetUnmarshaller(marshaller(p.format_input))
		}
	}
	log.Println("Press Ctrl-C to interrupt")
	p.Tasks.Add(&golib.NoopTask{golib.ExternalInterrupt(), "external interrupt"})
	p.Tasks.WaitAndExit()
}
