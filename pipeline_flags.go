package bitflow

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

type MarshallingFormat string
type EndpointType string

const (
	UndefinedEndpoint = EndpointType("")
	TcpEndpoint       = EndpointType("tcp")
	TcpListenEndpoint = EndpointType("listen")
	FileEndpoint      = EndpointType("file")
	StdEndpoint       = EndpointType("std")
	EmptyEndpoint     = EndpointType("empty")

	UndefinedFormat = MarshallingFormat("")
	TextFormat      = MarshallingFormat("text")
	CsvFormat       = MarshallingFormat("csv")
	BinaryFormat    = MarshallingFormat("bin")

	tcp_download_retry_interval = 1000 * time.Millisecond
	tcp_dial_timeout            = 2000 * time.Millisecond
)

var (
	allFormatsMap = map[MarshallingFormat]bool{
		TextFormat:   true,
		CsvFormat:    true,
		BinaryFormat: true,
	}

	stdTransportTarget = "-"
	binaryFileSuffix   = ".bin"
)

var DefaultEndpointFactory = EndpointFactory{
	FlagOutputFilesClean:   false,
	FlagIoBuffer:           4096,
	FlagTcpConnectionLimit: 0,
	FlagParallelHandler: ParallelSampleHandler{
		ParallelParsers: runtime.NumCPU(),
		BufferedSamples: 10000,
	},
	FlagFilesKeepAlive:        false,
	FlagInputFilesRobust:      false,
	FlagInputTcpAcceptLimit:   0,
	FlagTcpSourceDropErrors:   false,
	FlagOutputTcpListenBuffer: 0,
	FlagFilesAppend:           false,
	FlagFileVanishedCheck:     0,
}

func init() {
	DefaultEndpointFactory.Clear()
	RegisterDefaults(&DefaultEndpointFactory)
}

// EndpointFactory creates SampleSink and SampleSource instances for a SamplePipeline.
// It defines command line flags for configuring the objects it creates.
// All fields named Flag* are set by the according command line flags and evaluated in CreateInput() and CreateOutput().
// FlagInputs is not set by command line flags automatically.
// After flag.Parse(), those fields can be modified to override the command line flags defined by the user.
type EndpointFactory struct {
	// File input/output flags

	FlagInputFilesRobust  bool
	FlagOutputFilesClean  bool
	FlagIoBuffer          int
	FlagFilesKeepAlive    bool
	FlagFilesAppend       bool
	FlagFileVanishedCheck time.Duration

	// TCP input/output flags

	FlagOutputTcpListenBuffer uint
	FlagTcpConnectionLimit    uint
	FlagInputTcpAcceptLimit   uint
	FlagTcpSourceDropErrors   bool
	FlagTcpLogReceivedData    bool

	// Parallel marshalling/unmarshalling flags

	FlagParallelHandler ParallelSampleHandler

	// CustomDataSources can be filled by client code before EndpointFactory.CreateInput or similar
	// methods to allow creation of custom data sources. The map key is a short name of the data source
	// that can be used in URL endpoint descriptions. The parameter for the function will be
	// the URL path of the endpoint. Example: When registering a function with the key "http", the following
	// URL endpoint:
	//   http://localhost:5555/abc
	// will invoke the factory function with the parameter "localhost:5555/abc"
	CustomDataSources map[EndpointType]func(string) (SampleSource, error)

	// CustomDataSinks can be filled by client code before EndpointFactory.CreateOutput or similar
	// methods to allow creation of custom data sinks. See CustomDataSources for the meaning of the
	// map keys and values.
	CustomDataSinks map[EndpointType]func(string) (SampleProcessor, error)

	// CustomGeneralFlags, CustomInputFlags and CustomOutputFlags lets client code
	// register custom command line flags that configure aspects of endpoints created
	// through CustomDataSources and CustomDataSinks.
	CustomGeneralFlags []func(f *flag.FlagSet)
	CustomInputFlags   []func(f *flag.FlagSet)
	CustomOutputFlags  []func(f *flag.FlagSet)
}

func NewEndpointFactory() *EndpointFactory {
	factory := DefaultEndpointFactory
	factory.Clear()
	RegisterDefaults(&factory)
	return &factory
}

func (f *EndpointFactory) Clear() {
	f.CustomDataSources = make(map[EndpointType]func(string) (SampleSource, error))
	f.CustomDataSinks = make(map[EndpointType]func(string) (SampleProcessor, error))
	f.CustomGeneralFlags = nil
	f.CustomInputFlags = nil
	f.CustomOutputFlags = nil
}

func RegisterDefaults(factory *EndpointFactory) {
	RegisterConsoleBoxOutput(factory)
	RegisterEmptyInputOutput(factory)
}

func RegisterEmptyInputOutput(factory *EndpointFactory) {
	factory.CustomDataSinks[EmptyEndpoint] = func(string) (SampleProcessor, error) {
		return new(DroppingSampleProcessor), nil
	}
	factory.CustomDataSources[EmptyEndpoint] = func(string) (SampleSource, error) {
		return new(EmptySampleSource), nil
	}
}

func RegisterGolibFlags() {
	golib.RegisterFlags(golib.FlagsAll)
}

func (p *EndpointFactory) ParseParameters(params map[string]string) (err error) {
	copy := make(map[string]string, len(params))
	for key, val := range params {
		copy[key] = val
	}

	get := func(name string) string {
		if err != nil {
			return ""
		}
		res := params[name]
		delete(params, name)
		return res
	}
	boolParam := func(target *bool, name string) {
		if strVal := get(name); strVal != "" {
			*target, err = strconv.ParseBool(strVal)
		}
	}
	intParam := func(target *int, name string) {
		if strVal := get(name); strVal != "" {
			*target, err = strconv.Atoi(strVal)
		}
	}
	uintParam := func(target *uint, name string) {
		if strVal := get(name); strVal != "" {
			val, parseErr := strconv.ParseUint(strVal, 10, 64)
			err = parseErr
			*target = uint(val)
		}
	}
	durationParam := func(target *time.Duration, name string) {
		if strVal := get(name); strVal != "" {
			*target, err = time.ParseDuration(strVal)
		}
	}

	boolParam(&p.FlagOutputFilesClean, "files-clean")
	intParam(&p.FlagIoBuffer, "files-buf")
	uintParam(&p.FlagTcpConnectionLimit, "tcp-limit")
	boolParam(&p.FlagTcpLogReceivedData, "tcp-log-received")
	intParam(&p.FlagParallelHandler.ParallelParsers, "par")
	intParam(&p.FlagParallelHandler.BufferedSamples, "buf")
	boolParam(&p.FlagFilesKeepAlive, "files-keep-alive")
	boolParam(&p.FlagInputFilesRobust, "files-robust")
	uintParam(&p.FlagInputTcpAcceptLimit, "listen-limit")
	boolParam(&p.FlagTcpSourceDropErrors, "tcp-drop-err")
	uintParam(&p.FlagOutputTcpListenBuffer, "listen-buffer")
	boolParam(&p.FlagFilesAppend, "files-append")
	durationParam(&p.FlagFileVanishedCheck, "files-check-output")

	if len(copy) > 0 {
		return fmt.Errorf("Unexpected parameters for EndpointFactory: %v", copy)
	}
	return nil
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
	f.BoolVar(&p.FlagOutputFilesClean, "files-clean", p.FlagOutputFilesClean, "Delete all potential output files before writing.")
	f.IntVar(&p.FlagIoBuffer, "files-buf", p.FlagIoBuffer, "Size (byte) of buffered IO when reading/writing files.")

	// TCP
	f.UintVar(&p.FlagTcpConnectionLimit, "tcp-limit", p.FlagTcpConnectionLimit, "Limit number of TCP connections to accept/establish. Exit afterwards")

	// Parallel marshalling/unmarshalling
	f.IntVar(&p.FlagParallelHandler.ParallelParsers, "par", p.FlagParallelHandler.ParallelParsers, "Parallel goroutines used for (un)marshalling samples")
	f.IntVar(&p.FlagParallelHandler.BufferedSamples, "buf", p.FlagParallelHandler.BufferedSamples, "Number of samples buffered when (un)marshalling.")

	// Custom
	for _, factoryFunc := range p.CustomGeneralFlags {
		factoryFunc(f)
	}
}

// RegisterInputFlagsTo registers flags that configure aspects of data input.
func (p *EndpointFactory) RegisterInputFlagsTo(f *flag.FlagSet) {
	f.BoolVar(&p.FlagFilesKeepAlive, "files-keep-alive", p.FlagFilesKeepAlive, "Do not shut down after all files have been read. Useful in combination with -listen-buffer.")
	f.BoolVar(&p.FlagInputFilesRobust, "files-robust", p.FlagInputFilesRobust, "When encountering errors while reading files, print warnings instead of failing.")
	f.UintVar(&p.FlagInputTcpAcceptLimit, "listen-limit", p.FlagInputTcpAcceptLimit, "Limit number of simultaneous TCP connections accepted for incoming data.")
	f.BoolVar(&p.FlagTcpSourceDropErrors, "tcp-drop-err", p.FlagTcpSourceDropErrors, "Don't print errors when establishing active TCP input connection fails")
	for _, factoryFunc := range p.CustomInputFlags {
		factoryFunc(f)
	}
}

// RegisterOutputConfigFlagsTo registers flags that configure data outputs.
func (p *EndpointFactory) RegisterOutputFlagsTo(f *flag.FlagSet) {
	f.UintVar(&p.FlagOutputTcpListenBuffer, "listen-buffer", p.FlagOutputTcpListenBuffer, "When listening for outgoing connections, store a number of samples in a ring buffer that will be delivered first to all established connections.")
	f.BoolVar(&p.FlagFilesAppend, "files-append", p.FlagFilesAppend, "For file output, do no create new files by incrementing the suffix and append to existing files.")
	f.DurationVar(&p.FlagFileVanishedCheck, "files-check-output", p.FlagFileVanishedCheck, "For file output, check if the output file vanished or changed in regular intervals. Reopen the file in that case.")
	f.BoolVar(&p.FlagTcpLogReceivedData, "tcp-log-received", p.FlagTcpLogReceivedData, "For all TCP output connections, log received data, which is usually not expected.")
	for _, factoryFunc := range p.CustomOutputFlags {
		factoryFunc(f)
	}
}

// Writer returns an instance of SampleReader, configured by the values stored in the EndpointFactory.
func (p *EndpointFactory) Reader(um Unmarshaller) SampleReader {
	return SampleReader{
		ParallelSampleHandler: p.FlagParallelHandler,
		Unmarshaller:          um,
	}
}

// CreateInput creates a SampleSource object based on the given input endpoint descriptions
// and the configuration flags in the EndpointFactory.
func (p *EndpointFactory) CreateInput(inputs ...string) (SampleSource, error) {
	var result SampleSource
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
					PrintErrors:   !p.FlagTcpSourceDropErrors,
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
					FileNames: []string{endpoint.Target},
					IoBuffer:  p.FlagIoBuffer,
					Robust:    p.FlagInputFilesRobust,
					KeepAlive: p.FlagFilesKeepAlive,
				}
				source.Reader = reader
				result = source
			default:
				if factory, ok := p.CustomDataSources[endpoint.Type]; ok && endpoint.IsCustomType {
					var factoryErr error
					result, factoryErr = factory(endpoint.Target)
					if factoryErr != nil {
						return nil, fmt.Errorf("Error creating '%v' input: %v", endpoint.Type, factoryErr)
					}
				} else {
					return nil, errors.New("Unknown input endpoint type: " + string(endpoint.Type))
				}
			}
		} else {
			if inputType != endpoint.Type {
				return nil, fmt.Errorf("Please provide only one data source (Provided %v and %v)", inputType, endpoint.Type)
			}
			if endpoint.IsCustomType {
				return nil, fmt.Errorf("Cannot define multiple sources for custom input type '%v'", inputType)
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
				source.FileNames = append(source.FileNames, endpoint.Target)
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

// CreateInput creates a SampleSink object based on the given output endpoint description
// and the configuration flags in the EndpointFactory.
func (p *EndpointFactory) CreateOutput(output string) (SampleProcessor, error) {
	var resultSink SampleProcessor
	endpoint, err := ParseEndpointDescription(output, true)
	if err != nil {
		return nil, err
	}
	var marshallingSink *AbstractMarshallingSampleOutput
	marshaller := endpoint.OutputFormat().Marshaller()
	switch endpoint.Type {
	case StdEndpoint:
		sink := NewConsoleSink()
		marshallingSink = &sink.AbstractMarshallingSampleOutput
		if txt, ok := marshaller.(TextMarshaller); ok {
			txt.AssumeStdout = true
			marshaller = txt
		} else if txt, ok := marshaller.(*TextMarshaller); ok {
			txt.AssumeStdout = true
		}
		resultSink = sink
	case FileEndpoint:
		sink := &FileSink{
			Filename:          endpoint.Target,
			IoBuffer:          p.FlagIoBuffer,
			CleanFiles:        p.FlagOutputFilesClean,
			Append:            p.FlagFilesAppend,
			VanishedFileCheck: p.FlagFileVanishedCheck,
		}
		marshallingSink = &sink.AbstractMarshallingSampleOutput
		resultSink = sink
	case TcpEndpoint:
		sink := &TCPSink{
			Endpoint:    endpoint.Target,
			DialTimeout: tcp_dial_timeout,
		}
		sink.TcpConnLimit = p.FlagTcpConnectionLimit
		if p.FlagTcpLogReceivedData {
			sink.LogReceivedTraffic = log.ErrorLevel
		}
		marshallingSink = &sink.AbstractMarshallingSampleOutput
		resultSink = sink
	case TcpListenEndpoint:
		sink := &TCPListenerSink{
			Endpoint:        endpoint.Target,
			BufferedSamples: p.FlagOutputTcpListenBuffer,
		}
		sink.TcpConnLimit = p.FlagTcpConnectionLimit
		if p.FlagTcpLogReceivedData {
			sink.LogReceivedTraffic = log.ErrorLevel
		}
		marshallingSink = &sink.AbstractMarshallingSampleOutput
		resultSink = sink
	default:
		if factory, ok := p.CustomDataSinks[endpoint.Type]; ok && endpoint.IsCustomType {
			var factoryErr error
			resultSink, factoryErr = factory(endpoint.Target)
			if factoryErr != nil {
				return nil, fmt.Errorf("Error creating '%v' output: %v", endpoint.Type, factoryErr)
			}
		} else {
			return nil, errors.New("Unknown output endpoint type: " + string(endpoint.Type))
		}
	}
	if marshallingSink != nil {
		marshallingSink.SetMarshaller(marshaller)
		marshallingSink.Writer = p.Writer()
	}
	return resultSink, nil
}

// IsConsoleOutput returns true if the given processor will output to the standard output when started.
func IsConsoleOutput(sink SampleProcessor) bool {
	writer, ok1 := sink.(*WriterSink)
	_, ok2 := sink.(*ConsoleBoxSink)
	return (ok1 && writer.Output == os.Stdout) || ok2
}

// EndpointDescription describes a data endpoint, regardless of the data direction
// (input or output).
type EndpointDescription struct {
	Format       MarshallingFormat
	Type         EndpointType
	IsCustomType bool
	Target       string
}

// Unmarshaller creates an Unmarshaller object that is able to read data from the
// described endpoint.
func (e EndpointDescription) Unmarshaller() Unmarshaller {
	// The nil Unmarshaller makes the SampleSource implementations auto-detect the format.
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
	default:
		if e.IsCustomType {
			return UndefinedFormat
		} else {
			panic(fmt.Sprintf("Unknown endpoint type: %v", e.Type))
		}
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
		// This can occur with custom endpoints, where the Format is set as UndefinedFormat
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
		// Special case: Correct the default output transport type for standard output to ConsoleBoxEndpoint
		if err == nil && isOutput {
			if guessed.Target == stdTransportTarget && guessed.Format == UndefinedFormat {
				guessed.Type = ConsoleBoxEndpoint
				guessed.IsCustomType = true
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
			case StdEndpoint:
				if target != stdTransportTarget {
					err = fmt.Errorf("Transport '%v' can only be defined with target '%v'", part, stdTransportTarget)
					return
				}
				res.Type = StdEndpoint
			default:
				res.IsCustomType = true
				res.Type = EndpointType(part)
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
	if res.IsCustomType && res.Format != UndefinedFormat {
		err = fmt.Errorf("Cannot define the data format for transport '%v'", res.Type)
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
				var err golib.MultiError
				err.Add(err1)
				err.Add(err2)
				return UndefinedEndpoint, fmt.Errorf("Not a filename and not a valid TCP endpoint: %v (%v)", target, err.NilOrError())
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
