package sample

import (
	"errors"
	"flag"
	"os"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/golib"
)

const (
	tcp_download_retry_interval = 1000 * time.Millisecond
	input_formats               = "(a=auto, c=CSV, b=binary)"
	output_formats              = "(t=text, c=CSV, b=binary)"
)

func init() {
	// Configure logging output
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.StampMilli,
	})
	golib.Log = log.StandardLogger()
}

func marshaller(format string) Marshaller {
	switch format {
	case "t":
		return new(TextMarshaller)
	case "c":
		return new(CsvMarshaller)
	case "b":
		return new(BinaryMarshaller)
	default:
		log.WithField("format", format).Fatalln("Illegal data input fromat, must be one of", output_formats)
		return nil
	}
}

func unmarshaller(format string) Unmarshaller {
	switch format {
	case "a":
		return nil
	case "c":
		return new(CsvMarshaller)
	case "b":
		return new(BinaryMarshaller)
	default:
		log.WithField("format", format).Fatalln("Illegal data input fromat, must be one of", input_formats)
		return nil
	}
}

type CmdSamplePipeline struct {
	SamplePipeline
	Tasks             *golib.TaskGroup
	ReadSampleHandler ReadSampleHandler

	read_console           bool
	read_tcp_listen        string
	read_tcp_download      golib.StringSlice
	read_files             golib.StringSlice
	sink_console           bool
	sink_connect           golib.StringSlice
	sink_listen            golib.StringSlice
	sink_file              string
	format_input           string
	format_console         string
	format_connect         string
	format_listen          string
	format_file            string
	tcp_conn_limit         uint
	tcp_listen_limit       uint
	tcp_drop_active_errors bool
	handler                ParallelSampleHandler
	io_buf                 int
}

// Must be called before flag.Parse()
func (p *CmdSamplePipeline) ParseFlags() {
	flag.StringVar(&p.format_input, "i", "a", "Force data input format, default is auto-detect, one of "+input_formats)
	flag.BoolVar(&p.read_console, "C", false, "Data source: read from stdin")
	flag.Var(&p.read_files, "F", "Data source: read from file(s)")
	flag.StringVar(&p.read_tcp_listen, "L", "", "Data source: receive samples by accepting a TCP connection")
	flag.Var(&p.read_tcp_download, "D", "Data source: receive samples by connecting to remote endpoint(s)")

	flag.Var(&fileRegexValue{p}, "FR", "File Regex: Input files matching regex, previous -F parameter is used as start directory")
	flag.UintVar(&p.tcp_listen_limit, "Llimit", 0, "Limit number of simultaneous TCP connections accepted through -L.")
	flag.UintVar(&p.tcp_conn_limit, "tcp_limit", 0, "Limit number of TCP connections to accept/establish. Exit afterwards")
	flag.BoolVar(&p.tcp_drop_active_errors, "tcp_drop_err", false, "Dont print errors when establishing actie TCP connection (for sink/source) fails.")

	flag.IntVar(&p.handler.ParallelParsers, "par", runtime.NumCPU(), "Parallel goroutines used for (un)marshalling samples")
	flag.IntVar(&p.handler.BufferedSamples, "buf", 10000, "Number of samples buffered when (un)marshalling.")
	flag.IntVar(&p.io_buf, "io_buf", 4096, "Size (byte) of buffered IO when writing files.")

	flag.BoolVar(&p.sink_console, "p", false, "Data sink: print to stdout")
	flag.StringVar(&p.format_console, "pf", "t", "Data format for console output, one of "+output_formats)
	flag.StringVar(&p.sink_file, "f", "", "Data sink: write data to file")
	flag.StringVar(&p.format_file, "ff", "b", "Data format for file output, one of "+output_formats)
	flag.Var(&p.sink_connect, "s", "Data sink: send data to specified TCP endpoint")
	flag.StringVar(&p.format_connect, "sf", "b", "Data format for TCP output, one of "+output_formats)
	flag.Var(&p.sink_listen, "l", "Data sink: accept TCP connections for sending out data")
	flag.StringVar(&p.format_listen, "lf", "b", "Data format for TCP server output, one of "+output_formats)
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
		log.Fatalln("Error applying -FR flag:", err)
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
	reader := SampleReader{
		ParallelSampleHandler: p.handler,
		Handler:               p.ReadSampleHandler,
		Unmarshaller:          unmarshaller(p.format_input),
	}
	if p.read_console {
		p.SetSource(&ConsoleSource{Reader: reader})
	}
	if p.read_tcp_listen != "" {
		source := NewTcpListenerSource(p.read_tcp_listen, reader)
		source.SimultaneousConnections = p.tcp_listen_limit
		source.TcpConnLimit = p.tcp_conn_limit
		p.SetSource(source)
	}
	if len(p.read_tcp_download) > 0 {
		source := &TCPSource{
			RemoteAddrs:   p.read_tcp_download,
			RetryInterval: tcp_download_retry_interval,
			Reader:        reader,
			PrintErrors:   !p.tcp_drop_active_errors,
		}
		source.TcpConnLimit = p.tcp_conn_limit
		p.SetSource(source)
	}
	if len(p.read_files) > 0 {
		p.SetSource(&FileSource{Filenames: p.read_files, Reader: reader, IoBuffer: p.io_buf})
	}
	if p.Source == nil {
		log.Warnln("No data source provided, no data will be received or generated.")
		p.Source = new(EmptyMetricSource)
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
	for _, endpoint := range p.sink_connect {
		sink := &TCPSink{Endpoint: endpoint, PrintErrors: !p.tcp_drop_active_errors}
		sink.SetMarshaller(marshaller(p.format_connect))
		sink.Writer = writer
		sink.TcpConnLimit = p.tcp_conn_limit
		sinks = append(sinks, sink)
	}
	for _, endpoint := range p.sink_listen {
		sink := NewTcpListenerSink(endpoint)
		sink.SetMarshaller(marshaller(p.format_listen))
		sink.Writer = writer
		sink.TcpConnLimit = p.tcp_conn_limit
		sinks = append(sinks, sink)
	}
	if p.sink_file != "" {
		sink := &FileSink{Filename: p.sink_file, IoBuffer: p.io_buf}
		sink.SetMarshaller(marshaller(p.format_file))
		sink.Writer = writer
		sinks = append(sinks, sink)
	}
	if len(sinks) == 0 {
		log.Warnln("No data sinks selected, data will not be output anywhere.")
	}
	p.Sink = sinks
}

// p.Tasks can be filled with additional Tasks before calling this
func (p *CmdSamplePipeline) StartAndWait() int {
	p.Construct(p.Tasks)
	log.Println("Press Ctrl-C to interrupt")
	p.Tasks.Add(&golib.NoopTask{
		Chan:        golib.ExternalInterrupt(),
		Description: "external interrupt",
	})
	return p.Tasks.PrintWaitAndStop()
}
