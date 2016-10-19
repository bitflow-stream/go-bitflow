package data2go

import (
	"errors"
	"flag"
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

// CmdSamplePipeline is an extension for SamplePipeline that defines many command line flags
// and additional methods for controlling the pipeline. It is basically a helper type to be
// reused in different main packages.
//
// The sequence of operations on a CmdSamplePipeline should follow the following example:
//   // ... Define additional flags using the "flag" package (Optional)
//   var p sample.CmdSamplePipeline
//   p.ParseFlags()
//   flag.Parse()
//   defer golib.ProfileCpu()() // Optional
//   p.SetSource(customSource) // Optional
//   p.ReadSampleHandler = customHandler // Optional
//   p.Init()
//   p.Tasks.Add(customTask) // Optional
//   os.Exit(p.StartAndWait())
//
type CmdSamplePipeline struct {
	// SamplePipeline in the CmdSamplePipeline should not be accessed directly,
	// except for the Add method. The Sink and Source fields should only be read,
	// because they are set in the Init() method.
	SamplePipeline

	// Tasks will be initialized in Init() and should be accessed before
	// calling StartAndWait(). It can be used to add additional golib.Task instances
	// that should be started along with the pipeline in StartAndWait().
	// The pipeline tasks will be added in StartAndWait(), so the additional tasks
	// will be started before the pipeline.
	Tasks *golib.TaskGroup

	// ReadSampleHandler will be set to the Handler field in the SampleReader that is created in Init().
	// It can be used to modify Samples and Headers directly after they are read
	// by the SampleReader, e.g. directly after reading a Header from file, or directly
	// after receiving a Sample over a TCP connection. The main purpose of this
	// is that the ReadSampleHandler receives a string representation of the data
	// source, which can be used as a tag in the Samples. By default, the data source is not
	// stored in the Samples and this information will be lost one the Sample enters the pipeline.
	//
	// The data source string differs depending on the MetricSource used. The FileSource will
	// use the file name, while the TCPSource will use the remote TCP endpoint.
	//
	// Must be set before calling Init().
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
	robust_files           bool
	clean_files            bool
	handler                ParallelSampleHandler
	io_buf                 int
}

// ParseFlags registers all flags for the receiving CmdSamplePipeline with the
// global CommandLine object. The flags configure many aspects of the pipeline,
// including data source, data sink, performance parameters, debug parmeters and
// other behavior parameters. See the help texts for more information on the available
// parameters.
//
// Must be called before flag.Parse().
func (p *CmdSamplePipeline) ParseFlags() {
	flag.StringVar(&p.format_input, "i", "a", "Force data input format, default is auto-detect, one of "+input_formats)
	flag.BoolVar(&p.read_console, "C", false, "Data source: read from stdin")
	flag.Var(&p.read_files, "F", "Data source: read from file(s)")
	flag.StringVar(&p.read_tcp_listen, "L", "", "Data source: receive samples by accepting a TCP connection")
	flag.Var(&p.read_tcp_download, "D", "Data source: receive samples by connecting to remote endpoint(s)")

	flag.Var(&fileRegexValue{p}, "FR", "File Regex: Input files matching regex, previous -F parameter is used as start directory")
	flag.UintVar(&p.tcp_listen_limit, "Llimit", 0, "Limit number of simultaneous TCP connections accepted through -L.")
	flag.UintVar(&p.tcp_conn_limit, "tcp_limit", 0, "Limit number of TCP connections to accept/establish. Exit afterwards")
	flag.BoolVar(&p.tcp_drop_active_errors, "tcp_drop_err", false, "Don't print errors when establishing actie TCP connection (for sink/source) fails")
	flag.BoolVar(&p.robust_files, "robust", false, "Only print warnings for errors when reading files")
	flag.BoolVar(&p.clean_files, "clean", false, "Delete all potential output files before writing")

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

// SetSource allows external main packages to set their own MetricSource instances
// in the receiving CmdSamplePipeline. If this is used, none of the command line
// flags that otherwise define the data source can be used, as only one data source
// is allowed.
//
// Must be called before Init().
func (p *CmdSamplePipeline) SetSource(source MetricSource) {
	if p.Source != nil {
		log.Fatalln("Please provide only one data source")
	}
	p.Source = source
}

// Init initialized the receiving CmdSamplePipeline by creating the Tasks field,
// evaluating all command line flags and configurations, and setting the Source
// and Sink fields in the contained SamplePipeline.
//
// The only things to do before calling Init() is to set the ReadSampleHandler field
// and use the SetSource method, as both are evaluated in Init(). Both are optional.
//
// Init must be called before accessing p.Tasks and before calling StartAndWait().
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
		p.SetSource(&FileSource{
			Filenames: p.read_files,
			Reader:    reader,
			IoBuffer:  p.io_buf,
			Robust:    p.robust_files,
		})
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
		sink := &FileSink{Filename: p.sink_file, IoBuffer: p.io_buf, CleanFiles: p.clean_files}
		sink.SetMarshaller(marshaller(p.format_file))
		sink.Writer = writer
		sinks = append(sinks, sink)
	}
	if len(sinks) == 0 {
		log.Warnln("No data sinks selected, data will not be output anywhere.")
	}
	p.Sink = sinks
}

// StartAndWait constructs the pipeline and starts it. It blocks until the pipeline
// is finished.
//
// An additional golib.Task is stared along with the pipeline, which listens
// for the Ctrl-C user external interrupt and makes the pipeline stoppable cleanly
// when the user wants to.
//
// p.Tasks can be filled with additional Tasks before calling this.
//
// StartAndWait returns the number of errors that occured in the pipeline or in
// additional tasks added to the TaskGroup.
func (p *CmdSamplePipeline) StartAndWait() int {
	p.Construct(p.Tasks)
	log.Debugln("Press Ctrl-C to interrupt")
	p.Tasks.Add(&golib.NoopTask{
		Chan:        golib.ExternalInterrupt(),
		Description: "external interrupt",
	})
	return p.Tasks.PrintWaitAndStop()
}

// ============= Helper types and methods =============

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
