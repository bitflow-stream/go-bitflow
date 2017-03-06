package bitflow

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
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
	UndefinedEndpoint  = EndpointType("")
	TcpEndpoint        = EndpointType("tcp")
	TcpListenEndpoint  = EndpointType("listen")
	FileEndpoint       = EndpointType("file")
	StdEndpoint        = EndpointType("std")
	ConsoleBoxEndpoint = EndpointType("box")

	UndefinedFormat = MarshallingFormat("")
	TextFormat      = MarshallingFormat("text")
	CsvFormat       = MarshallingFormat("csv")
	BinaryFormat    = MarshallingFormat("bin")

	tcp_download_retry_interval = 1000 * time.Millisecond
	tcp_dial_timeout            = 2000 * time.Millisecond
)

var (
	OutputFormats = []MarshallingFormat{TextFormat, CsvFormat, BinaryFormat}
	AllTransports = []EndpointType{TcpEndpoint, TcpListenEndpoint, FileEndpoint, StdEndpoint, ConsoleBoxEndpoint}

	allFormatsMap = map[MarshallingFormat]bool{
		TextFormat:   true,
		CsvFormat:    true,
		BinaryFormat: true,
	}

	ConsoleBoxSettings = gotermBox.CliLogBox{
		NoUtf8:        false,
		LogLines:      10,
		MessageBuffer: 500,
	}
	ConsoleBoxUpdateInterval    = 500 * time.Millisecond
	ConsoleBoxMinUpdateInterval = 50 * time.Millisecond

	stdTransportTarget = "-"
	binaryFileSuffix   = ".bin"
)

// EndpointFactory creates MetricSink and MetricSource instances for a SamplePipeline.
// It defines command line flags for configuring the objects it creates.
// All fields named Flag* are set by the according command line flags and evaluated in CreateInput() and CreateOutput().
// FlagInputs is not set by command line flags automatically.
// After flag.Parse(), those fields can be modified to override the command line flags defined by the user.
type EndpointFactory struct {
	// File input/output flags

	FlagInputFilesRobust bool
	FlagOutputFilesClean bool
	FlagIoBuffer         int
	FlagFilesKeepAlive   bool

	// TCP input/output flags

	FlagOutputTcpListenBuffer uint
	FlagTcpConnectionLimit    uint
	FlagInputTcpAcceptLimit   uint
	FlagTcpDropErrors         bool

	// Parallel marshalling/unmarshalling flags

	FlagParallelHandler ParallelSampleHandler

	// Console box flags

	ConsoleBoxNoImmediateScreenUpdate bool

	// testmode is a flag used by tests to suppress initialization routines
	// that are not testable. It is a hack to keep the EndpointFactory easy to use
	// while making it testable.
	testmode bool
}

func RegisterGolibFlags() {
	golib.RegisterFlags(golib.FlagsAll & ^golib.FlagsOFL)
}

// RegisterConfigFlags registers all flags to the global CommandLine object.
func (p *EndpointFactory) RegisterFlags() {
	p.RegisterGeneralFlagsTo(flag.CommandLine)
	p.RegisterInputFlagsTo(flag.CommandLine)
	p.RegisterOutputFlagsTo(flag.CommandLine)
}

// RegisterGeneralFlagsTo registers flags that configure different aspects of both
// data input and data output. These flags affect to both performance and functionality of
// TCP, file and std I/O.
func (p *EndpointFactory) RegisterGeneralFlagsTo(f *flag.FlagSet) {
	// Files
	f.BoolVar(&p.FlagOutputFilesClean, "files-clean", false, "Delete all potential output files before writing.")
	f.IntVar(&p.FlagIoBuffer, "files-buf", 4096, "Size (byte) of buffered IO when reading/writing files.")

	// TCP
	f.UintVar(&p.FlagTcpConnectionLimit, "tcp-limit", 0, "Limit number of TCP connections to accept/establish. Exit afterwards")
	f.BoolVar(&p.FlagTcpDropErrors, "tcp-drop-err", false, "Don't print errors when establishing active TCP connection (for sink/source) fails")

	// Parallel marshalling/unmarshalling
	f.IntVar(&p.FlagParallelHandler.ParallelParsers, "par", runtime.NumCPU(), "Parallel goroutines used for (un)marshalling samples")
	f.IntVar(&p.FlagParallelHandler.BufferedSamples, "buf", 10000, "Number of samples buffered when (un)marshalling.")
}

// RegisterInputFlagsTo registers flags that configure aspects of data input.
func (p *EndpointFactory) RegisterInputFlagsTo(f *flag.FlagSet) {
	f.BoolVar(&p.FlagFilesKeepAlive, "files-keep-alive", false, "Do not shut down after all files have been read. Useful in combination with -listen-buffer.")
	f.BoolVar(&p.FlagInputFilesRobust, "files-robust", false, "When encountering errors while reading files, print warnings instead of failing.")
	f.UintVar(&p.FlagInputTcpAcceptLimit, "listen-limit", 0, "Limit number of simultaneous TCP connections accepted for incoming data.")
}

// RegisterOutputConfigFlagsTo registers flags that configure data outputs.
func (p *EndpointFactory) RegisterOutputFlagsTo(f *flag.FlagSet) {
	f.UintVar(&p.FlagOutputTcpListenBuffer, "listen-buffer", 0, "When listening for outgoing connections, store a number of samples in a ring buffer that will be delivered first to all established connections.")
	f.BoolVar(&p.ConsoleBoxNoImmediateScreenUpdate, "slow-screen-updates", false, fmt.Sprintf("For console box output, don't update the screen on every sample, but only in intervals of %v", ConsoleBoxUpdateInterval))
}

// Writer returns an instance of SampleReader, configured by the values stored in the EndpointFactory.
func (p *EndpointFactory) Reader(um Unmarshaller) SampleReader {
	return SampleReader{
		ParallelSampleHandler: p.FlagParallelHandler,
		Unmarshaller:          um,
	}
}

// CreateInput creates a MetricSource object based on the given input endpoint descriptions
// and the configuration flags in the EndpointFactory.
func (p *EndpointFactory) CreateInput(inputs ...string) (UnmarshallingMetricSource, error) {
	var result UnmarshallingMetricSource
	inputType := UndefinedEndpoint
	for _, input := range inputs {
		endpoint, err := ParseEndpointDescription(input, false)
		if err != nil {
			return nil, err
		}
		if endpoint.Format != UndefinedFormat {
			return nil, fmt.Errorf("Format cannot be specified for data input: %v", input)
		}
		if result == nil {
			reader := p.Reader(endpoint.Unmarshaller())
			inputType = endpoint.Type
			switch endpoint.Type {
			case StdEndpoint:
				source := NewConsoleSource()
				source.Reader = reader
				result = source
			case TcpEndpoint:
				source := &TCPSource{
					RemoteAddrs:   []string{endpoint.Target},
					PrintErrors:   !p.FlagTcpDropErrors,
					RetryInterval: tcp_download_retry_interval,
					DialTimeout:   tcp_dial_timeout,
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
					KeepAlive: p.FlagFilesKeepAlive,
				}
				source.Reader = reader
				result = source
			default:
				return nil, errors.New("Unknown input endpoint type: " + string(endpoint.Type))
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
				return nil, errors.New("Unknown endpoint type: " + string(endpoint.Type))
			}
		}
	}
	return result, nil
}

// Writer returns an instance of SampleWriter, configured by the values stored in the EndpointFactory.
func (p *EndpointFactory) Writer() SampleWriter {
	return SampleWriter{p.FlagParallelHandler}
}

// CreateInput creates a MetricSink object based on the given output endpoint description
// and the configuration flags in the EndpointFactory.
func (p *EndpointFactory) CreateOutput(output string) (MetricSink, error) {
	var resultSink MetricSink
	endpoint, err := ParseEndpointDescription(output, true)
	if err != nil {
		return nil, err
	}
	var marshallingSink *AbstractMarshallingMetricSink
	marshaller := endpoint.OutputFormat().Marshaller()
	switch endpoint.Type {
	case StdEndpoint:
		sink := NewConsoleSink()
		marshallingSink = &sink.AbstractMarshallingMetricSink
		if txt, ok := marshaller.(TextMarshaller); ok {
			txt.AssumeStdout = true
			marshaller = txt
		} else if txt, ok := marshaller.(*TextMarshaller); ok {
			txt.AssumeStdout = true
		}
		resultSink = sink
	case ConsoleBoxEndpoint:
		sink := &ConsoleBoxSink{
			CliLogBoxTask: gotermBox.CliLogBoxTask{
				CliLogBox:         ConsoleBoxSettings,
				UpdateInterval:    ConsoleBoxUpdateInterval,
				MinUpdateInterval: ConsoleBoxMinUpdateInterval,
			},
			ImmediateScreenUpdate: !p.ConsoleBoxNoImmediateScreenUpdate,
		}
		if !p.testmode {
			sink.Init()
		}
		resultSink = sink
	case FileEndpoint:
		sink := &FileSink{
			Filename:   endpoint.Target,
			IoBuffer:   p.FlagIoBuffer,
			CleanFiles: p.FlagOutputFilesClean,
		}
		marshallingSink = &sink.AbstractMarshallingMetricSink
		resultSink = sink
	case TcpEndpoint:
		sink := &TCPSink{
			Endpoint:    endpoint.Target,
			PrintErrors: !p.FlagTcpDropErrors,
			DialTimeout: tcp_dial_timeout,
		}
		sink.TcpConnLimit = p.FlagTcpConnectionLimit
		marshallingSink = &sink.AbstractMarshallingMetricSink
		resultSink = sink
	case TcpListenEndpoint:
		sink := &TCPListenerSink{
			Endpoint:        endpoint.Target,
			BufferedSamples: p.FlagOutputTcpListenBuffer,
		}
		sink.TcpConnLimit = p.FlagTcpConnectionLimit
		marshallingSink = &sink.AbstractMarshallingMetricSink
		resultSink = sink
	default:
		return nil, errors.New("Unknown output endpoint type: " + string(endpoint.Type))
	}
	if marshallingSink != nil {
		marshallingSink.SetMarshaller(marshaller)
		marshallingSink.Writer = p.Writer()
	}
	return resultSink, nil
}

// IsConsoleOutput returns true if the given MetricSink will output to the standard output when started.
func IsConsoleOutput(sink MetricSink) bool {
	writer, ok1 := sink.(*WriterSink)
	_, ok2 := sink.(*ConsoleBoxSink)
	return (ok1 && writer.Output == os.Stdout) || ok2
}

// EndpointDescription describes a data endpoint, regardless of the data direction
// (input or output).
type EndpointDescription struct {
	Format MarshallingFormat
	Type   EndpointType
	Target string
}

// Unmarshaller creates an Unmarshaller object that is able to read data from the
// described endpoint.
func (e EndpointDescription) Unmarshaller() Unmarshaller {
	// The nil Unmarshaller makes the MetricSource implementations auto-detect the format.
	return nil
}

// OutputFormat returns the MarshallingFormat that should be used when sending
// data to the described endpoint.
func (e EndpointDescription) OutputFormat() MarshallingFormat {
	format := e.Format
	if format == UndefinedFormat {
		format = e.DefaultOutputFormat()
	}
	return format
}

// DefaultOutputFormat returns the default MarshallingFormat that should be used when sending
// data to the described endpoint, if no format is specified by the user.
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
	case ConsoleBoxEndpoint:
		return UndefinedFormat
	default:
		panic("Unknown endpoint type: " + e.Type)
	}
}

// Marshaller returns a Marshaller object that is able to marshall data for sending
// it to the described endpoint.
func (format MarshallingFormat) Marshaller() Marshaller {
	switch format {
	case TextFormat:
		return TextMarshaller{}
	case CsvFormat:
		return CsvMarshaller{}
	case BinaryFormat:
		return BinaryMarshaller{}
	default:
		// This can occurr with ConsoleBoxEndpoint, where the Format is parsed as UndefinedFormat
		return nil
	}
}

// ParseEndpointDescription parses the given string to an EndpointDescription object.
// The string can be one of two forms: the URL-style description will be parsed by
// ParseUrlEndpointDescription, other descriptions will be parsed by GuessEndpointDescription.
func ParseEndpointDescription(endpoint string, isOutput bool) (EndpointDescription, error) {
	if strings.Contains(endpoint, "://") {
		return ParseUrlEndpointDescription(endpoint)
	} else {
		guessed, err := GuessEndpointDescription(endpoint)
		//qs Correct the default output transport type for standard output to ConsoleBoxEndpoint
		if err == nil && isOutput {
			if guessed.Target == "-" && guessed.Format == UndefinedFormat {
				guessed.Type = ConsoleBoxEndpoint
			}
		}
		return guessed, err
	}
}

// ParseUrlEndpointDescription parses the endpoint string as a URL endpoint description.
// It has the form:
//   format+transport://target
//
// One of the format and transport parts must be specified, optionally both.
// If one of format or transport is missing, it will be guessed.
// The order does not matter. The 'target' part must not be empty.
func ParseUrlEndpointDescription(endpoint string) (res EndpointDescription, err error) {
	urlParts := strings.SplitN(endpoint, "://", 2)
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
			case StdEndpoint, ConsoleBoxEndpoint:
				if target != stdTransportTarget {
					err = fmt.Errorf("Transport '%v' can only be defined with target '%v'", part, stdTransportTarget)
					return
				}
				res.Type = EndpointType(part)
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
	if res.Type == ConsoleBoxEndpoint && res.Format != UndefinedFormat {
		err = fmt.Errorf("Cannot define the format for transport '%v'", ConsoleBoxEndpoint)
	}
	return
}

// GuessEndpointDescription guesses the transport type and format of the given endpoint target.
// See GuessEndpointType for details.
func GuessEndpointDescription(endpoint string) (res EndpointDescription, err error) {
	res.Target = endpoint
	res.Type, err = GuessEndpointType(endpoint)
	return
}

// GuessEndpointType guesses the EndpointType for the given target.
// Three forms of are recognized for the target:
//  - A host:port pair indicates an active TCP endpoint
//  - A :port pair (without the host part, but with the colon) indicates a passive TCP endpoint listening on the given port.
//  - The hyphen '-' is interpreted as standard input/output.
//  - All other targets are treated as file names.
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
			if strings.Contains(target, ":") || !IsValidFilename(target) {
				return UndefinedEndpoint, fmt.Errorf("Not a filename and not a valid TCP endpoint: %v", target)
			}
			typ = FileEndpoint
		}
	}
	log.Debugf("Guessed transport type of %v: %v", target, typ)
	return typ, nil
}

// IsValidFilename tries to infer in a system-independent way, if the given path is a valid file name.
func IsValidFilename(path string) bool {
	_, err := os.Stat(path)
	switch err := err.(type) {
	case *os.PathError:
		return err == nil || err.Err == nil || err.Err.Error() != "invalid argument"
	default:
		return true
	}
}
