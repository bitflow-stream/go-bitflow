package bitflow

import (
	"errors"
	"flag"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/golib"
	"github.com/antongulenko/golib/gotermBox"
)

const (
	tcp_download_retry_interval = 1000 * time.Millisecond
	input_formats               = "(a=auto, c=CSV, b=binary)"
	output_formats              = "(t=text, c=CSV, b=binary)"
)

// CmdSamplePipeline is an extension for SamplePipeline that defines many command line flags
// and additional methods for controlling the pipeline. It is basically a helper type to be
// reused in different main packages. All fields named Flag* are set by the according command
// line flags and evaluated in Init(). After flag.Parse(), those fields can be modified
// to override the command line flags defined by the user.
//
// The sequence of operations on a CmdSamplePipeline should follow the following example:
//   // ... Define additional flags using the "flag" package (Optional)
//   var p sample.CmdSamplePipeline
//   p.ParseFlags()
//   flag.Parse()
//   defer golib.ProfileCpu()() // Optional
//   p.SetSource(customSource) // Optional
//   p.ReadSampleHandler = customHandler // Optional
//   // ... Modify p.Flag* values // Optional
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

	// Input flags

	FlagInputFormat    string
	FlagInputConsole   bool
	FlagInputFiles     golib.StringSlice
	FlagInputTcpListen string
	FlagInputTcp       golib.StringSlice

	// File input/output flags

	FlagInputFilesRobust bool
	FlagOutputFilesClean bool

	// TCP input/output flags

	FlagOutputTcpListenBuffer uint
	FlagTcpConnectionLimit    uint
	FlagInputTcpAcceptLimit   uint
	FlagTcpDropErrors         bool

	// Parallel marshalling/unmarshalling flags

	FlagParallelHandler ParallelSampleHandler
	FlagIoBuffer        int

	// Output flags

	FlagOutputConsole         bool
	FlagOutputBox             bool
	FlagOutputTcp             golib.StringSlice
	FlagOutputTcpListen       golib.StringSlice
	FlagOutputFile            string
	FlagOutputConsoleFormat   string
	FlagOutputTcpFormat       string
	FlagOutputTcpListenFormat string
	FlagOutputFileFormat      string
}

// ParseAllFlags calls ParseFlags without suppressing any available flags.
func (p *CmdSamplePipeline) RegisterAllFlags() {
	p.RegisterFlags(nil)
}

// ParseFlags registers command-line flags for the receiving CmdSamplePipeline with the
// global CommandLine object. The flags configure many aspects of the pipeline,
// including data source, data sink, performance parameters, debug parmeters and
// other behavior parameters. See the help texts for more information on the available
// parameters.
//
// The suppressFlags parameter can be filled with flags, that should not be parsed.
// This allows individual main-packages to refine the available flags.
//
// Must be called before flag.Parse().
func (p *CmdSamplePipeline) RegisterFlags(suppressFlags map[string]bool) {
	var f flag.FlagSet

	// Input
	f.StringVar(&p.FlagInputFormat, "i", "a", "Force data input format, default is auto-detect, one of "+input_formats)
	f.BoolVar(&p.FlagInputConsole, "C", false, "Data source: read from stdin")
	f.Var(&p.FlagInputFiles, "F", "Data source: read from file(s)")
	f.StringVar(&p.FlagInputTcpListen, "L", "", "Data source: receive samples by accepting a TCP connection")
	f.Var(&p.FlagInputTcp, "D", "Data source: receive samples by connecting to remote endpoint(s)")

	// File input
	f.Var(&fileRegexValue{p}, "FR", "File Regex: Input files matching regex, previous -F parameter is used as start directory")
	f.BoolVar(&p.FlagInputFilesRobust, "robust", false, "Only print warnings for errors when reading files")

	// File output
	f.BoolVar(&p.FlagOutputFilesClean, "clean", false, "Delete all potential output files before writing")

	// TCP input/output
	f.UintVar(&p.FlagOutputTcpListenBuffer, "lbuf", 0, "For -l, buffer a number of samples that will be delivered first to all established connections.")
	f.UintVar(&p.FlagInputTcpAcceptLimit, "Llimit", 0, "Limit number of simultaneous TCP connections accepted through -L.")
	f.UintVar(&p.FlagTcpConnectionLimit, "tcp_limit", 0, "Limit number of TCP connections to accept/establish. Exit afterwards")
	f.BoolVar(&p.FlagTcpDropErrors, "tcp_drop_err", false, "Don't print errors when establishing actie TCP connection (for sink/source) fails")

	// Parallel marshalling/unmarshalling
	f.IntVar(&p.FlagParallelHandler.ParallelParsers, "par", runtime.NumCPU(), "Parallel goroutines used for (un)marshalling samples")
	f.IntVar(&p.FlagParallelHandler.BufferedSamples, "buf", 10000, "Number of samples buffered when (un)marshalling.")
	f.IntVar(&p.FlagIoBuffer, "io_buf", 4096, "Size (byte) of buffered IO when writing files.")

	// Output
	f.BoolVar(&p.FlagOutputBox, "p", false, "Data sink: print to box on stdout. Not possible with -c.")
	f.BoolVar(&p.FlagOutputConsole, "c", false, "Data sink: print to stdout")
	f.StringVar(&p.FlagOutputConsoleFormat, "cf", "t", "Data format for console output, one of "+output_formats)
	f.StringVar(&p.FlagOutputFile, "f", "", "Data sink: write data to file")
	f.StringVar(&p.FlagOutputFileFormat, "ff", "b", "Data format for file output, one of "+output_formats)
	f.Var(&p.FlagOutputTcp, "s", "Data sink: send data to specified TCP endpoint")
	f.StringVar(&p.FlagOutputTcpFormat, "sf", "b", "Data format for TCP output, one of "+output_formats)
	f.Var(&p.FlagOutputTcpListen, "l", "Data sink: accept TCP connections for sending out data")
	f.StringVar(&p.FlagOutputTcpListenFormat, "lf", "b", "Data format for TCP server output, one of "+output_formats)

	f.VisitAll(func(f *flag.Flag) {
		if !suppressFlags[f.Name] {
			flag.Var(f.Value, f.Name, f.Usage)
		}
	})
}

// HasOutputFlag returns true, if at least one data output flag is defined in the
// receiving CmdSamplePipeline. If false is returned, calling Init() will print
// a warning since the data will not be output anywhere.
func (p *CmdSamplePipeline) HasOutputFlag() bool {
	return p.FlagOutputBox || p.FlagOutputConsole || p.FlagOutputFile != "" || len(p.FlagOutputTcp) > 0 || len(p.FlagOutputTcpListen) > 0
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
// p.ParseFlags() should also be called.
//
// Init must be called before accessing p.Tasks and before calling StartAndWait().
//
// See the documentation of the CmdSamplePipeline type.
func (p *CmdSamplePipeline) Init() {
	p.Tasks = golib.NewTaskGroup()

	// ====== Initialize sink(s) and tasks
	var sinks AggregateSink
	writer := SampleWriter{p.FlagParallelHandler}
	if p.FlagOutputBox {
		if p.FlagOutputConsole {
			log.Fatalln("Cannot specify both -c and -p")
		}
		sink := &ConsoleBoxSink{
			CliLogBox: gotermBox.CliLogBox{
				NoUtf8:        false,
				LogLines:      10,
				MessageBuffer: 500,
			},
			UpdateInterval: 500 * time.Millisecond,
		}
		sink.Init()
		sinks = append(sinks, sink)
	}
	if p.FlagOutputConsole {
		sink := new(ConsoleSink)
		m := marshaller(p.FlagOutputConsoleFormat)
		if txt, ok := m.(*TextMarshaller); ok {
			txt.AssumeStdout = true
		}
		sink.SetMarshaller(m)
		sink.Writer = writer
		sinks = append(sinks, sink)
	}
	for _, endpoint := range p.FlagOutputTcp {
		sink := &TCPSink{Endpoint: endpoint, PrintErrors: !p.FlagTcpDropErrors}
		sink.SetMarshaller(marshaller(p.FlagOutputTcpFormat))
		sink.Writer = writer
		sink.TcpConnLimit = p.FlagTcpConnectionLimit
		sinks = append(sinks, sink)
	}
	for _, endpoint := range p.FlagOutputTcpListen {
		sink := NewTcpListenerSink(endpoint, p.FlagOutputTcpListenBuffer)
		sink.SetMarshaller(marshaller(p.FlagOutputTcpListenFormat))
		sink.Writer = writer
		sink.TcpConnLimit = p.FlagTcpConnectionLimit
		sinks = append(sinks, sink)
	}
	if p.FlagOutputFile != "" {
		sink := &FileSink{Filename: p.FlagOutputFile, IoBuffer: p.FlagIoBuffer, CleanFiles: p.FlagOutputFilesClean}
		sink.SetMarshaller(marshaller(p.FlagOutputFileFormat))
		sink.Writer = writer
		sinks = append(sinks, sink)
	}
	if len(sinks) == 0 {
		log.Warnln("No data sinks selected, data will not be output anywhere.")
	}
	p.Sink = sinks

	// ====== Initialize source(s)
	reader := SampleReader{
		ParallelSampleHandler: p.FlagParallelHandler,
		Handler:               p.ReadSampleHandler,
		Unmarshaller:          unmarshaller(p.FlagInputFormat),
	}
	if p.FlagInputConsole {
		p.SetSource(&ConsoleSource{Reader: reader})
	}
	if p.FlagInputTcpListen != "" {
		source := NewTcpListenerSource(p.FlagInputTcpListen, reader)
		source.SimultaneousConnections = p.FlagInputTcpAcceptLimit
		source.TcpConnLimit = p.FlagTcpConnectionLimit
		p.SetSource(source)
	}
	if len(p.FlagInputTcp) > 0 {
		source := &TCPSource{
			RemoteAddrs:   p.FlagInputTcp,
			RetryInterval: tcp_download_retry_interval,
			Reader:        reader,
			PrintErrors:   !p.FlagTcpDropErrors,
		}
		source.TcpConnLimit = p.FlagTcpConnectionLimit
		p.SetSource(source)
	}
	if len(p.FlagInputFiles) > 0 {
		p.SetSource(&FileSource{
			Filenames: p.FlagInputFiles,
			Reader:    reader,
			IoBuffer:  p.FlagIoBuffer,
			Robust:    p.FlagInputFilesRobust,
		})
	}
	if p.Source == nil {
		log.Warnln("No data source provided, no data will be received or generated.")
		if len(sinks) == 0 && len(p.Processors) == 0 {
			log.Fatalln("No tasks defined")
		}
		p.Source = new(EmptyMetricSource)
	}
}

// StartAndWait constructs the pipeline and starts it. It blocks until the pipeline
// is finished.
//
// An additional golib.Task is started along with the pipeline, which listens
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
	if len(f.pipeline.FlagInputFiles) == 0 {
		return errors.New("-FR flag must appear after -F")
	}
	dir := f.pipeline.FlagInputFiles[0]
	f.pipeline.FlagInputFiles = f.pipeline.FlagInputFiles[1:]
	files, err := ListMatchingFiles(dir, param)
	if err != nil {
		log.Fatalln("Error applying -FR flag:", err)
		return err
	}
	f.pipeline.FlagInputFiles = append(f.pipeline.FlagInputFiles, files...)
	return nil
}
