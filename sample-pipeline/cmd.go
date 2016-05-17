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

type CmdSamplePipeline struct {
	SamplePipeline
	Tasks *golib.TaskGroup

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

// Must be called before flag.Parse()
func (p *CmdSamplePipeline) ParseFlags() {
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

// Must be called before p.Init()
func (p *CmdSamplePipeline) SetSource(source sample.MetricSource) {
	if p.Source != nil {
		log.Fatalln("Please provide only one data source")
	}
	p.Source = source
}

// Must be called before accessing p.Tasks and calling p.StartAndWait()
func (p *CmdSamplePipeline) Init() {
	p.Tasks = golib.NewTaskGroup()

	// ====== Initialize source(s)
	if p.read_console {
		p.SetSource(new(sample.ConsoleSource))
	}
	if p.read_tcp_listen != "" {
		p.SetSource(sample.NewTcpListenerSource(p.read_tcp_listen))
	}
	if p.read_tcp_download != "" {
		p.SetSource(&sample.TCPSource{
			RemoteAddr:    p.read_tcp_download,
			RetryInterval: tcp_download_retry_interval,
		})
	}
	if p.read_file != "" {
		p.SetSource(&sample.FileSource{
			FileTransport: sample.FileTransport{Filename: p.read_file}})
	}
	if p.Source == nil {
		log.Println("No data source provided, no data will be received or generated.")
	} else {
		if umSource, ok := p.Source.(sample.UnmarshallingMetricSource); ok {
			umSource.SetUnmarshaller(marshaller(p.format_input))
		}
	}

	// ====== Initialize sink(s) and tasks
	var sinks sample.AggregateSink
	if p.sink_console {
		sink := new(sample.ConsoleSink)
		sink.SetMarshaller(marshaller(p.format_console))
		sinks = append(sinks, sink)
	}
	if p.sink_connect != "" {
		sink := &sample.TCPSink{Endpoint: p.sink_connect}
		sink.SetMarshaller(marshaller(p.format_connect))
		sinks = append(sinks, sink)
	}
	if p.sink_listen != "" {
		sink := sample.NewTcpListenerSink(p.sink_listen)
		sink.SetMarshaller(marshaller(p.format_listen))
		sinks = append(sinks, sink)
	}
	if p.sink_file != "" {
		sink := &sample.FileSink{FileTransport: sample.FileTransport{Filename: p.sink_file}}
		sink.SetMarshaller(marshaller(p.format_file))
		sinks = append(sinks, sink)
	}
	if len(sinks) == 0 {
		log.Println("No data sinks selected, data will not be output anywhere.")
	}
	p.Sink = sinks
}

// p.Tasks can be filled with additional Tasks before calling this
func (p *CmdSamplePipeline) StartAndWait() {
	p.Construct(p.Tasks)
	log.Println("Press Ctrl-C to interrupt")
	p.Tasks.Add(&golib.NoopTask{golib.ExternalInterrupt(), "external interrupt"})
	p.Tasks.WaitAndExit()
}
