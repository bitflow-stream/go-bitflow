package sample

import (
	"errors"
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

func marshaller(format string) MetricMarshaller {
	switch format {
	case "c":
		return new(CsvMarshaller)
	case "b":
		return new(BinaryMarshaller)
	case "t":
		return new(TextMarshaller)
	default:
		log.Fatalf("Illegal data fromat '%v', must be one of %v\n", format, supported_formats)
		return nil
	}
}

type CmdSamplePipeline struct {
	SamplePipeline
	Tasks          *golib.TaskGroup
	SampleReadHook SampleReadHook

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
	tcp_conn_limit    uint
	handler           ParallelSampleHandler
}

// Must be called before flag.Parse()
func (p *CmdSamplePipeline) ParseFlags() {
	flag.StringVar(&p.format_input, "i", "", "Data source format, default depends on selected source, one of "+supported_formats)
	flag.BoolVar(&p.read_console, "C", false, "Data source: read from stdin")
	flag.Var(&p.read_files, "F", "Data source: read from file(s)")
	flag.StringVar(&p.read_tcp_listen, "L", "", "Data source: receive samples by accepting a TCP connection")
	flag.StringVar(&p.read_tcp_download, "D", "", "Data source: receive samples by connecting to remote endpoint")

	flag.Var(&fileRegexValue{p}, "FR", "File Regex: Input files matching regex, previous -F parameter is used as start directory")
	flag.UintVar(&p.tcp_conn_limit, "tcp_limit", 0, "Limit number of TCP connections to accept/establish. Exit afterwards")

	flag.IntVar(&p.handler.ParallelParsers, "par", runtime.NumCPU(), "Parallel goroutines used for (un)marshalling samples")
	flag.IntVar(&p.handler.BufferedSamples, "buf", 10000, "Number of samples buffered when (un)marshalling")
	flag.IntVar(&p.handler.IoBuffer, "io_buf", 4096, "Size (byte) of buffered IO when (un)marshalling")

	flag.BoolVar(&p.sink_console, "p", false, "Data sink: print to stdout")
	flag.StringVar(&p.format_console, "pf", "t", "Data format for console output, one of "+supported_formats)
	flag.StringVar(&p.sink_file, "f", "", "Data sink: write data to file")
	flag.StringVar(&p.format_file, "ff", "c", "Data format for file output, one of "+supported_formats)
	flag.StringVar(&p.sink_connect, "s", "", "Data sink: send data to specified TCP endpoint")
	flag.StringVar(&p.format_connect, "sf", "b", "Data format for TCP output, one of "+supported_formats)
	flag.StringVar(&p.sink_listen, "l", "", "Data sink: accept TCP connections for sending out data")
	flag.StringVar(&p.format_listen, "lf", "b", "Data format for TCP server output, one of "+supported_formats)
}

type fileRegexValue struct {
	pipeline *CmdSamplePipeline
}

func (f *fileRegexValue) String() string {
	return ""
}

func (f *fileRegexValue) Set(param string) error {
	if len(f.pipeline.read_files) == 0 {
		return errors.New("-FR flag must appear after -F")
	}
	dir := f.pipeline.read_files[0]
	f.pipeline.read_files = f.pipeline.read_files[1:]
	files, err := ListMatchingFiles(dir, param)
	if err != nil {
		log.Fatalf("Error applying -FR flag: %v\n", err)
		return err
	}
	f.pipeline.read_files = append(f.pipeline.read_files, files...)
	return nil
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
		ParallelSampleHandler: p.handler,
		ReadHook:              p.SampleReadHook,
	}
	if p.read_console {
		p.SetSource(&ConsoleSource{Reader: reader})
		unmarshaller = default_console_input
	}
	if p.read_tcp_listen != "" {
		source := NewTcpListenerSource(p.read_tcp_listen, reader)
		source.TcpConnLimit = p.tcp_conn_limit
		p.SetSource(source)
		unmarshaller = default_tcp_input
	}
	if p.read_tcp_download != "" {
		source := &TCPSource{
			RemoteAddr:    p.read_tcp_download,
			RetryInterval: tcp_download_retry_interval,
			Reader:        reader,
		}
		source.TcpConnLimit = p.tcp_conn_limit
		p.SetSource(source)
		unmarshaller = default_tcp_input
	}
	if len(p.read_files) > 0 {
		p.SetSource(&FileSource{Filenames: p.read_files, Reader: reader})
		unmarshaller = default_file_input
	}
	if p.Source == nil {
		log.Println("No data source provided, no data will be received or generated.")
		p.Source = new(EmptyMetricSource)
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
	writer := SampleWriter{p.handler}
	if p.sink_console {
		sink := new(ConsoleSink)
		m := marshaller(p.format_console)
		if txt, ok := m.(*TextMarshaller); ok {
			txt.AssumeStdout = true
		}
		sink.SetMarshaller(m)
		sink.Writer = writer
		sinks = append(sinks, sink)
	}
	if p.sink_connect != "" {
		sink := &TCPSink{Endpoint: p.sink_connect}
		sink.SetMarshaller(marshaller(p.format_connect))
		sink.Writer = writer
		sink.TcpConnLimit = p.tcp_conn_limit
		sinks = append(sinks, sink)
	}
	if p.sink_listen != "" {
		sink := NewTcpListenerSink(p.sink_listen)
		sink.SetMarshaller(marshaller(p.format_listen))
		sink.Writer = writer
		sink.TcpConnLimit = p.tcp_conn_limit
		sinks = append(sinks, sink)
	}
	if p.sink_file != "" {
		sink := &FileSink{Filename: p.sink_file}
		sink.SetMarshaller(marshaller(p.format_file))
		sink.Writer = writer
		sinks = append(sinks, sink)
	}
	if len(sinks) == 0 {
		log.Println("No data sinks selected, data will not be output anywhere.")
	}
	p.Sink = sinks
}

// p.Tasks can be filled with additional Tasks before calling this
func (p *CmdSamplePipeline) StartAndWait() int {
	p.Construct(p.Tasks)
	log.Println("Press Ctrl-C to interrupt")
	p.Tasks.Add(&golib.NoopTask{golib.ExternalInterrupt(), "external interrupt"})
	return p.Tasks.PrintWaitAndStop()
}
