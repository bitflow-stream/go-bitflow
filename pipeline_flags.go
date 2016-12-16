package bitflow

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/golib"
	"github.com/antongulenko/golib/gotermBox"
)

type MarshallingFormat string
type EndpointType string

const (
	UndefinedEndpoint = EndpointType("")
	TcpEndpoint       = EndpointType("tcp")
	TcpListenEndpoint = EndpointType("listen")
	FileEndpoint      = EndpointType("file")
	StdEndpoint       = EndpointType("std")

	UndefinedFormat = MarshallingFormat("")
	TextFormat      = MarshallingFormat("text")
	CsvFormat       = MarshallingFormat("csv")
	BinaryFormat    = MarshallingFormat("bin")
)

var (
	OutputFormats = []MarshallingFormat{TextFormat, CsvFormat, BinaryFormat}
	AllTransports = []EndpointType{TcpEndpoint, TcpListenEndpoint, FileEndpoint, StdEndpoint}

	allFormatsMap = map[MarshallingFormat]bool{
		TextFormat:   true,
		CsvFormat:    true,
		BinaryFormat: true,
	}

	stdTransportTarget = "-"
	binaryFileSuffix   = ".bin"
)

// EndpointFactory creates MetricSink and MetricSource instances for a CmdSamplePipeline.
// Ir defines command line flags for configuring the instances it creates.
type EndpointFactory struct {
	// File input/output flags

	FlagInputFilesRobust bool
	FlagOutputFilesClean bool
	FlagIoBuffer         int

	// TCP input/output flags

	FlagOutputTcpListenBuffer uint
	FlagTcpConnectionLimit    uint
	FlagInputTcpAcceptLimit   uint
	FlagTcpDropErrors         bool

	// Parallel marshalling/unmarshalling flags

	FlagParallelHandler ParallelSampleHandler

	// Output

	FlagOutputBox bool
	FlagOutputs   golib.StringSlice

	// Input

	FlagInputs golib.StringSlice
}

// ParseAllFlags calls ParseFlags without suppressing any available flags.
func (p *EndpointFactory) RegisterAllFlags() {
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
func (p *EndpointFactory) RegisterFlags(suppressFlags map[string]bool) {
	var f flag.FlagSet

	// Files
	f.BoolVar(&p.FlagInputFilesRobust, "robust", false, "When encountering errors while reading files, print warnings instead of failing.")
	f.BoolVar(&p.FlagOutputFilesClean, "clean", false, "Delete all potential output files before writing")
	f.IntVar(&p.FlagIoBuffer, "io_buf", 4096, "Size (byte) of buffered IO when reading/writing files.")

	// TCP
	f.UintVar(&p.FlagOutputTcpListenBuffer, "lbuf", 0, "For -l, buffer a number of samples that will be delivered first to all established connections.")
	f.UintVar(&p.FlagInputTcpAcceptLimit, "Llimit", 0, "Limit number of simultaneous TCP connections accepted through -L.")
	f.UintVar(&p.FlagTcpConnectionLimit, "tcp_limit", 0, "Limit number of TCP connections to accept/establish. Exit afterwards")
	f.BoolVar(&p.FlagTcpDropErrors, "tcp_drop_err", false, "Don't print errors when establishing actie TCP connection (for sink/source) fails")

	// Parallel marshalling/unmarshalling
	f.IntVar(&p.FlagParallelHandler.ParallelParsers, "par", runtime.NumCPU(), "Parallel goroutines used for (un)marshalling samples")
	f.IntVar(&p.FlagParallelHandler.BufferedSamples, "buf", 10000, "Number of samples buffered when (un)marshalling.")

	// Output
	f.BoolVar(&p.FlagOutputBox, "p", false, "Display samples in a box on the command line")
	f.Var(&p.FlagOutputs, "o", "Data sink(s) for outputting data")

	f.VisitAll(func(f *flag.Flag) {
		if !suppressFlags[f.Name] {
			flag.Var(f.Value, f.Name, f.Usage)
		}
	})
}

// HasOutputFlag returns true, if at least one data output flag is defined in the
// receiving CmdSamplePipeline. If false is returned, calling Init() will print
// a warning since the data will not be output anywhere.
func (p *EndpointFactory) HasOutputFlag() bool {
	return p.FlagOutputBox || len(p.FlagOutputs) > 0
}

func (p *EndpointFactory) CreateInput(handler ReadSampleHandler) (MetricSource, error) {
	var result MetricSource
	inputType := UndefinedEndpoint
	for _, input := range p.FlagInputs {
		endpoint, err := ParseEndpointDescription(input)
		if err != nil {
			return nil, err
		}
		if endpoint.Format != UndefinedFormat {
			return nil, fmt.Errorf("Format cannot be specified for data input: %v", input)
		}
		if result == nil {
			reader := SampleReader{
				ParallelSampleHandler: p.FlagParallelHandler,
				Handler:               handler,
				Unmarshaller:          endpoint.Unmarshaller(),
			}
			inputType = endpoint.Type
			switch endpoint.Type {
			case StdEndpoint:
				source := new(ConsoleSource)
				source.Reader = reader
				result = source
			case TcpEndpoint:
				source := &TCPSource{
					RemoteAddrs:   []string{endpoint.Target},
					RetryInterval: tcp_download_retry_interval,
					PrintErrors:   !p.FlagTcpDropErrors,
				}
				source.TcpConnLimit = p.FlagTcpConnectionLimit
				source.Reader = reader
				result = source
			case TcpListenEndpoint:
				source := NewTcpListenerSource(endpoint.Target)
				source.SimultaneousConnections = p.FlagInputTcpAcceptLimit
				source.TcpConnLimit = p.FlagTcpConnectionLimit
				source.Reader = reader
				result = source
			case FileEndpoint:
				source := &FileSource{
					Filenames: []string{endpoint.Target},
					IoBuffer:  p.FlagIoBuffer,
					Robust:    p.FlagInputFilesRobust,
				}
				source.Reader = reader
				result = source
			default:
				log.Println("HHHHHHHHHHHHHHHHHHH", endpoint)
				panic("Unknown endpoint type: " + string(endpoint.Type))
			}
		} else {
			if inputType != endpoint.Type {
				return nil, fmt.Errorf("Please provide only one data source (Provided %v and %v)", inputType, endpoint.Type)
			}
			switch endpoint.Type {
			case StdEndpoint:
				return nil, errors.New("Cannot read from stdin multiple times")
			case TcpListenEndpoint:
				return nil, errors.New("Cannot listen for input on multiple TCP ports")
			case TcpEndpoint:
				source := result.(*TCPSource)
				source.RemoteAddrs = append(source.RemoteAddrs, endpoint.Target)
			case FileEndpoint:
				source := result.(*FileSource)
				source.Filenames = append(source.Filenames, endpoint.Target)
			default:
				panic("Unknown endpoint type: " + endpoint.Type)
			}
		}
	}
	return result, nil
}

func (p *EndpointFactory) CreateOutput() (AggregateSink, error) {
	var sinks AggregateSink
	haveConsoleOutput := false
	for _, output := range p.FlagOutputs {
		endpoint, err := ParseEndpointDescription(output)
		if err != nil {
			return nil, err
		}
		var marshallingSink *AbstractMarshallingMetricSink
		marshaller := endpoint.OutputFormat().Marshaller()
		switch endpoint.Type {
		case StdEndpoint:
			if haveConsoleOutput {
				return nil, errors.New("Cannot define multiple outputs to stdout")
			}
			haveConsoleOutput = true
			sink := new(ConsoleSink)
			marshallingSink = &sink.AbstractMarshallingMetricSink
			sinks = append(sinks, sink)
			if txt, ok := marshaller.(*TextMarshaller); ok {
				txt.AssumeStdout = true
			}
		case FileEndpoint:
			sink := &FileSink{
				Filename:   endpoint.Target,
				IoBuffer:   p.FlagIoBuffer,
				CleanFiles: p.FlagOutputFilesClean,
			}
			marshallingSink = &sink.AbstractMarshallingMetricSink
			sinks = append(sinks, sink)
		case TcpEndpoint:
			sink := &TCPSink{
				Endpoint:    endpoint.Target,
				PrintErrors: !p.FlagTcpDropErrors,
			}
			sink.TcpConnLimit = p.FlagTcpConnectionLimit
			marshallingSink = &sink.AbstractMarshallingMetricSink
			sinks = append(sinks, sink)
		case TcpListenEndpoint:
			sink := NewTcpListenerSink(endpoint.Target, p.FlagOutputTcpListenBuffer)
			sink.TcpConnLimit = p.FlagTcpConnectionLimit
			marshallingSink = &sink.AbstractMarshallingMetricSink
			sinks = append(sinks, sink)
		default:
			panic("Unknown endpoint type: " + endpoint.Type)
		}
		marshallingSink.SetMarshaller(marshaller)
		marshallingSink.Writer = SampleWriter{p.FlagParallelHandler}
	}
	if p.FlagOutputBox {
		if haveConsoleOutput {
			return nil, errors.New("Cannot define multiple outputs to stdout")
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
	return sinks, nil
}

type EndpointDescription struct {
	Format MarshallingFormat
	Type   EndpointType
	Target string
}

func (e EndpointDescription) Unmarshaller() Unmarshaller {
	// The nil Unmarshaller makes the MetricSource implementations auto-detect the format.
	return nil
}

func (e EndpointDescription) OutputFormat() MarshallingFormat {
	format := e.Format
	if format == UndefinedFormat {
		format = e.DefaultOutputFormat()
	}
	return format
}

func (e EndpointDescription) DefaultOutputFormat() MarshallingFormat {
	switch e.Type {
	case TcpEndpoint, TcpListenEndpoint:
		return BinaryFormat
	case FileEndpoint:
		if strings.HasSuffix(e.Target, binaryFileSuffix) {
			return BinaryFormat
		}
		return CsvFormat
	case StdEndpoint:
		return TextFormat
	default:
		panic("Unknown endpoint type: " + e.Type)
	}
}

func (format MarshallingFormat) Marshaller() Marshaller {
	switch format {
	case TextFormat:
		return new(TextMarshaller)
	case CsvFormat:
		return new(CsvMarshaller)
	case BinaryFormat:
		return new(BinaryMarshaller)
	default:
		log.WithField("format", format).Fatalln("Illegal data output fromat, must be one of:", OutputFormats)
		return nil
	}
}

func ParseEndpointDescription(endpoint string) (EndpointDescription, error) {
	if strings.Contains(endpoint, "://") {
		return ParseUrlEndpointDescription(endpoint)
	} else {
		return GuessEndpointDescription(endpoint)
	}
}

func ParseUrlEndpointDescription(endpoint string) (res EndpointDescription, err error) {
	urlParts := strings.Split(endpoint, "://")
	if len(urlParts) != 2 || urlParts[0] == "" || urlParts[1] == "" {
		err = fmt.Errorf("Invalid URL endpoint: %v", endpoint)
		return
	}
	target := urlParts[1]
	res.Target = target
	for _, part := range strings.Split(urlParts[0], "+") {
		if allFormatsMap[MarshallingFormat(part)] {
			if res.Format != "" {
				err = fmt.Errorf("Multiple formats defined in: %v", endpoint)
				return
			}
			res.Format = MarshallingFormat(part)
		} else {
			if res.Type != UndefinedEndpoint {
				err = fmt.Errorf("Multiple transport types defined: %v", endpoint)
				return
			}
			switch EndpointType(part) {
			case TcpEndpoint:
				res.Type = TcpEndpoint
			case TcpListenEndpoint:
				res.Type = TcpListenEndpoint
			case FileEndpoint:
				res.Type = FileEndpoint
			case StdEndpoint:
				if target != stdTransportTarget {
					err = fmt.Errorf("Transport '%v' can only be defined with target '%v'", StdEndpoint, stdTransportTarget)
					return
				}
				res.Type = StdEndpoint
			default:
				err = fmt.Errorf("Illegal transport type: %v", part)
				return
			}
		}
	}
	if res.Type == UndefinedEndpoint {
		var guessErr error
		res.Type, guessErr = GuessEndpointType(target)
		if guessErr != nil {
			err = guessErr
		}
	}
	return
}

func GuessEndpointDescription(endpoint string) (res EndpointDescription, err error) {
	res.Target = endpoint
	res.Type, err = GuessEndpointType(endpoint)
	return
}

func GuessEndpointType(target string) (EndpointType, error) {
	var typ EndpointType
	if target == "" {
		return UndefinedEndpoint, errors.New("Empty endpoint/file is not valid")
	} else if target == stdTransportTarget {
		typ = StdEndpoint
	} else {
		host, port, err1 := net.SplitHostPort(target)
		_, err2 := net.LookupPort("tcp", port)
		if err1 == nil && err2 == nil {
			if host == "" {
				typ = TcpListenEndpoint
			} else {
				typ = TcpEndpoint
			}
		} else {
			// TODO query if target would be a valid file name
			if strings.Contains(target, ":") {
				return UndefinedEndpoint, fmt.Errorf("Not a filename and not a valid TCP endpoint: %v", target)
			}
			typ = FileEndpoint
		}
	}
	log.Debugf("Guessed transport type of %v: %v", target, typ)
	return typ, nil
}
