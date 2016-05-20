package sample

import (
	"flag"
	"log"
	"runtime"
	"time"

	"github.com/antongulenko/golib"
)

const (
	tcp_download_retry_interval = 1000 * time.Millisecond
	supported_formats           = "(c=CSV, b=binary, t=text)"

	default_file_input    = "c"
	default_console_input = "c"
	default_tcp_input     = "b"
)

var (
	marshallers = map[string]MetricMarshaller{
		"c": new(CsvMarshaller),
		"b": new(BinaryMarshaller),
		"t": new(TextMarshaller),
	}
)

func marshaller(format string) MetricMarshaller {
	if marshaller, ok := marshallers[format]; !ok {
		log.Fatalf("Illegal data fromat '%v', must be one of %v\n", format, supported_formats)
		return nil
	} else {
		return marshaller
	}
}

type CmdSamplePipeline struct {
	SamplePipeline
	Tasks *golib.TaskGroup

	parallel_parsers int
	buffered_samples int
	write_buffer     int

	read_console      bool
	read_tcp_listen   string
	read_tcp_download string
	read_files        golib.StringSlice
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
	flag.StringVar(&p.format_input, "i", "", "Data source format, default depends on selected source, one of "+supported_formats)
	flag.BoolVar(&p.read_console, "C", false, "Data source: read from stdin")
	flag.Var(&p.read_files, "F", "Data source: read from file(s)")
	flag.StringVar(&p.read_tcp_listen, "L", "", "Data source: receive samples by accepting a TCP connection")
	flag.StringVar(&p.read_tcp_download, "D", "", "Data source: receive samples by connecting to remote endpoint")

	flag.IntVar(&p.parallel_parsers, "par", runtime.NumCPU(), "Parallel goroutines used for unmarshalling samples")
	flag.IntVar(&p.buffered_samples, "read_buf", 10000, "Number of samples buffered when reading")
	flag.IntVar(&p.write_buffer, "write_buf", 4096, "Size (byte) of buffered IO for all data sinks")

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
func (p *CmdSamplePipeline) SetSource(source MetricSource) {
	if p.Source != nil {
		log.Fatalln("Please provide only one data source")
	}
	p.Source = source
}

// Must be called before accessing p.Tasks and calling p.StartAndWait()
func (p *CmdSamplePipeline) Init() {
	p.Tasks = golib.NewTaskGroup()

	// ====== Initialize source(s)
	var unmarshaller string
	reader := SampleReader{
		ParallelParsers: p.parallel_parsers,
		BufferedSamples: p.buffered_samples,
	}
	if p.read_console {
		p.SetSource(&ConsoleSource{Reader: reader})
		unmarshaller = default_console_input
	}
	if p.read_tcp_listen != "" {
		p.SetSource(NewTcpListenerSource(p.read_tcp_listen, reader))
		unmarshaller = default_tcp_input
	}
	if p.read_tcp_download != "" {
		p.SetSource(&TCPSource{
			RemoteAddr:    p.read_tcp_download,
			RetryInterval: tcp_download_retry_interval,
			Reader:        reader,
		})
		unmarshaller = default_tcp_input
	}
	if len(p.read_files) > 0 {
		p.SetSource(&FileSource{Filenames: p.read_files, Reader: reader})
		unmarshaller = default_file_input
	}
	if p.Source == nil {
		log.Println("No data source provided, no data will be received or generated.")
	} else {
		if p.format_input != "" {
			unmarshaller = p.format_input
		}
		if umSource, ok := p.Source.(UnmarshallingMetricSource); ok {
			umSource.SetUnmarshaller(marshaller(unmarshaller))
		}
	}

	// ====== Initialize sink(s) and tasks
	var sinks AggregateSink
	if p.sink_console {
		sink := new(ConsoleSink)
		sink.WriteBuffer = p.write_buffer
		sink.SetMarshaller(marshaller(p.format_console))
		sinks = append(sinks, sink)
	}
	if p.sink_connect != "" {
		sink := &TCPSink{Endpoint: p.sink_connect}
		sink.WriteBuffer = p.write_buffer
		sink.SetMarshaller(marshaller(p.format_connect))
		sinks = append(sinks, sink)
	}
	if p.sink_listen != "" {
		sink := NewTcpListenerSink(p.sink_listen)
		sink.WriteBuffer = p.write_buffer
		sink.SetMarshaller(marshaller(p.format_listen))
		sinks = append(sinks, sink)
	}
	if p.sink_file != "" {
		sink := &FileSink{Filename: p.sink_file}
		sink.WriteBuffer = p.write_buffer
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
