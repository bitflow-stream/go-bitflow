package bitflow

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
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
	HttpEndpoint      = EndpointType("http")
	EmptyEndpoint     = EndpointType("empty")

	UndefinedFormat  = MarshallingFormat("")
	TextFormat       = MarshallingFormat("text")
	CsvFormat        = MarshallingFormat("csv")
	BinaryFormat     = MarshallingFormat("bin")
	PrometheusFormat = MarshallingFormat("prometheus")

	tcp_download_retry_interval = 1000 * time.Millisecond
	tcp_dial_timeout            = 2000 * time.Millisecond
)

var (
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
	FlagSourceTag string

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

	// Marshallers can be filled by client code before EndpointFactory.CreateOutput or similar
	// methods to allow custom marshalling formats in output files, network connections and so on.
	Marshallers map[MarshallingFormat]func() Marshaller

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
	f.Marshallers = make(map[MarshallingFormat]func() Marshaller)
	f.CustomGeneralFlags = nil
	f.CustomInputFlags = nil
	f.CustomOutputFlags = nil
}

func RegisterDefaults(factory *EndpointFactory) {
	RegisterBuiltinMarshallers(factory)
	RegisterConsoleBoxOutput(factory)
	RegisterEmptyInputOutput(factory)
	RegisterDynamicSource(factory)
}

func RegisterEmptyInputOutput(factory *EndpointFactory) {
	factory.CustomDataSinks[EmptyEndpoint] = func(string) (SampleProcessor, error) {
		return new(DroppingSampleProcessor), nil
	}
	factory.CustomDataSources[EmptyEndpoint] = func(string) (SampleSource, error) {
		return new(EmptySampleSource), nil
	}
}

func RegisterBuiltinMarshallers(factory *EndpointFactory) {
	factory.Marshallers[TextFormat] = func() Marshaller {
		return TextMarshaller{}
	}
	factory.Marshallers[CsvFormat] = func() Marshaller {
		return CsvMarshaller{}
	}
	factory.Marshallers[BinaryFormat] = func() Marshaller {
		return BinaryMarshaller{}
	}
	factory.Marshallers[PrometheusFormat] = func() Marshaller {
		return PrometheusMarshaller{}
	}
}

func (f *EndpointFactory) ParseParameters(params map[string]string) (err error) {
	get := func(name string) string {
		if err != nil {
			return ""
		}
		res := params[name]
		delete(params, name)
		return res
	}
	strParam := func(target *string, name string) {
		*target = get(name)
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

	strParam(&f.FlagSourceTag, "source-tag")
	boolParam(&f.FlagOutputFilesClean, "files-clean")
	intParam(&f.FlagIoBuffer, "files-buf")
	uintParam(&f.FlagTcpConnectionLimit, "tcp-limit")
	boolParam(&f.FlagTcpLogReceivedData, "tcp-log-received")
	intParam(&f.FlagParallelHandler.ParallelParsers, "par")
	intParam(&f.FlagParallelHandler.BufferedSamples, "buf")
	boolParam(&f.FlagFilesKeepAlive, "files-keep-alive")
	boolParam(&f.FlagInputFilesRobust, "files-robust")
	uintParam(&f.FlagInputTcpAcceptLimit, "listen-limit")
	boolParam(&f.FlagTcpSourceDropErrors, "tcp-drop-err")
	uintParam(&f.FlagOutputTcpListenBuffer, "listen-buffer")
	boolParam(&f.FlagFilesAppend, "files-append")
	durationParam(&f.FlagFileVanishedCheck, "files-check-output")

	if err == nil && len(params) > 0 {
		err = fmt.Errorf("Unexpected parameters for EndpointFactory: %v", params)
	}
	return
}

// RegisterConfigFlags registers all flags to the global CommandLine object.
func (f *EndpointFactory) RegisterFlags() {
	f.RegisterGeneralFlagsTo(flag.CommandLine)
	f.RegisterInputFlagsTo(flag.CommandLine)
	f.RegisterOutputFlagsTo(flag.CommandLine)
}

// RegisterGeneralFlagsTo registers flags that configure different aspects of both
// data input and data output. These flags affect to both performance and functionality of
// TCP, file and std I/O.
func (f *EndpointFactory) RegisterGeneralFlagsTo(fs *flag.FlagSet) {
	// Files
	fs.BoolVar(&f.FlagOutputFilesClean, "files-clean", f.FlagOutputFilesClean, "Delete all potential output files before writing.")
	fs.IntVar(&f.FlagIoBuffer, "files-buf", f.FlagIoBuffer, "Size (byte) of buffered IO when reading/writing files.")

	// TCP
	fs.UintVar(&f.FlagTcpConnectionLimit, "tcp-limit", f.FlagTcpConnectionLimit, "Limit number of TCP connections to accept/establish. Exit afterwards")

	// Parallel marshalling/unmarshalling
	fs.IntVar(&f.FlagParallelHandler.ParallelParsers, "par", f.FlagParallelHandler.ParallelParsers, "Parallel goroutines used for (un)marshalling samples")
	fs.IntVar(&f.FlagParallelHandler.BufferedSamples, "buf", f.FlagParallelHandler.BufferedSamples, "Number of samples buffered when (un)marshalling.")

	// Custom
	for _, factoryFunc := range f.CustomGeneralFlags {
		factoryFunc(fs)
	}
}

// RegisterInputFlagsTo registers flags that configure aspects of data input.
func (f *EndpointFactory) RegisterInputFlagsTo(fs *flag.FlagSet) {
	fs.StringVar(&f.FlagSourceTag, "source-tag", f.FlagSourceTag, "Add the data source (e.g. input file, TCP endpoint, ...) as the given tag to each read sample.")
	fs.BoolVar(&f.FlagFilesKeepAlive, "files-keep-alive", f.FlagFilesKeepAlive, "Do not shut down after all files have been read. Useful in combination with -listen-buffer.")
	fs.BoolVar(&f.FlagInputFilesRobust, "files-robust", f.FlagInputFilesRobust, "When encountering errors while reading files, print warnings instead of failing.")
	fs.UintVar(&f.FlagInputTcpAcceptLimit, "listen-limit", f.FlagInputTcpAcceptLimit, "Limit number of simultaneous TCP connections accepted for incoming data.")
	fs.BoolVar(&f.FlagTcpSourceDropErrors, "tcp-drop-err", f.FlagTcpSourceDropErrors, "Don't print errors when establishing active TCP input connection fails")
	for _, factoryFunc := range f.CustomInputFlags {
		factoryFunc(fs)
	}
}

// RegisterOutputConfigFlagsTo registers flags that configure data outputs.
func (f *EndpointFactory) RegisterOutputFlagsTo(fs *flag.FlagSet) {
	fs.UintVar(&f.FlagOutputTcpListenBuffer, "listen-buffer", f.FlagOutputTcpListenBuffer, "When listening for outgoing connections, store a number of samples in a ring buffer that will be delivered first to all established connections.")
	fs.BoolVar(&f.FlagFilesAppend, "files-append", f.FlagFilesAppend, "For file output, do no create new files by incrementing the suffix and append to existing files.")
	fs.DurationVar(&f.FlagFileVanishedCheck, "files-check-output", f.FlagFileVanishedCheck, "For file output, check if the output file vanished or changed in regular intervals. Reopen the file in that case.")
	fs.BoolVar(&f.FlagTcpLogReceivedData, "tcp-log-received", f.FlagTcpLogReceivedData, "For all TCP output connections, log received data, which is usually not expected.")
	for _, factoryFunc := range f.CustomOutputFlags {
		factoryFunc(fs)
	}
}

// Writer returns an instance of SampleReader, configured by the values stored in the EndpointFactory.
func (f *EndpointFactory) Reader(um Unmarshaller) SampleReader {
	return SampleReader{
		ParallelSampleHandler: f.FlagParallelHandler,
		Unmarshaller:          um,
	}
}

// CreateInput creates a SampleSource object based on the given input endpoint descriptions
// and the configuration flags in the EndpointFactory.
func (f *EndpointFactory) CreateInput(inputs ...string) (SampleSource, error) {
	var result SampleSource
	inputType := UndefinedEndpoint
	for _, input := range inputs {
		endpoint, err := f.ParseEndpointDescription(input, false)
		if err != nil {
			return nil, err
		}
		if endpoint.Format != UndefinedFormat {
			return nil, fmt.Errorf("Format cannot be specified for data input: %v", input)
		}
		if result == nil {
			reader := f.Reader(nil) // nil as Unmarshaller makes the SampleSource auto-detect the format
			if f.FlagSourceTag != "" {
				reader.Handler = sourceTagger(f.FlagSourceTag)
			}
			inputType = endpoint.Type
			switch endpoint.Type {
			case StdEndpoint:
				source := NewConsoleSource()
				source.Reader = reader
				result = source
			case TcpEndpoint, HttpEndpoint:
				source := &TCPSource{
					RemoteAddrs:   []string{endpoint.Target},
					PrintErrors:   !f.FlagTcpSourceDropErrors,
					RetryInterval: tcp_download_retry_interval,
					DialTimeout:   tcp_dial_timeout,
					UseHTTP:       endpoint.Type == HttpEndpoint,
				}
				source.TcpConnLimit = f.FlagTcpConnectionLimit
				source.Reader = reader
				result = source
			case TcpListenEndpoint:
				source := NewTcpListenerSource(endpoint.Target)
				source.SimultaneousConnections = f.FlagInputTcpAcceptLimit
				source.TcpConnLimit = f.FlagTcpConnectionLimit
				source.Reader = reader
				result = source
			case FileEndpoint:
				source := &FileSource{
					FileNames: []string{endpoint.Target},
					IoBuffer:  f.FlagIoBuffer,
					Robust:    f.FlagInputFilesRobust,
					KeepAlive: f.FlagFilesKeepAlive,
				}
				source.Reader = reader
				result = source
			default:
				if factory, ok := f.CustomDataSources[endpoint.Type]; ok && endpoint.IsCustomType {
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
			case TcpEndpoint, HttpEndpoint:
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
func (f *EndpointFactory) Writer() SampleWriter {
	return SampleWriter{f.FlagParallelHandler}
}

// CreateInput creates a SampleSink object based on the given output endpoint description
// and the configuration flags in the EndpointFactory.
func (f *EndpointFactory) CreateOutput(output string) (SampleProcessor, error) {
	var resultSink SampleProcessor
	endpoint, err := f.ParseEndpointDescription(output, true)
	if err != nil {
		return nil, err
	}
	var marshaller Marshaller
	if format := endpoint.OutputFormat(); format != UndefinedFormat {
		marshaller, err = f.CreateMarshaller(format)
		if err != nil {
			return nil, err
		}
	}
	var marshallingSink *AbstractMarshallingSampleOutput
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
			IoBuffer:          f.FlagIoBuffer,
			CleanFiles:        f.FlagOutputFilesClean,
			Append:            f.FlagFilesAppend,
			VanishedFileCheck: f.FlagFileVanishedCheck,
		}
		marshallingSink = &sink.AbstractMarshallingSampleOutput
		resultSink = sink
	case TcpEndpoint:
		sink := &TCPSink{
			Endpoint:    endpoint.Target,
			DialTimeout: tcp_dial_timeout,
		}
		sink.TcpConnLimit = f.FlagTcpConnectionLimit
		if f.FlagTcpLogReceivedData {
			sink.LogReceivedTraffic = log.ErrorLevel
		}
		marshallingSink = &sink.AbstractMarshallingSampleOutput
		resultSink = sink
	case TcpListenEndpoint:
		sink := &TCPListenerSink{
			Endpoint:        endpoint.Target,
			BufferedSamples: f.FlagOutputTcpListenBuffer,
		}
		sink.TcpConnLimit = f.FlagTcpConnectionLimit
		if f.FlagTcpLogReceivedData {
			sink.LogReceivedTraffic = log.ErrorLevel
		}
		marshallingSink = &sink.AbstractMarshallingSampleOutput
		resultSink = sink
	case HttpEndpoint:
		theUrl, err := url.Parse("http://" + endpoint.Target)
		if err != nil {
			return nil, err
		}
		sink := &HttpServerSink{
			Endpoint:        theUrl.Host,
			RootPathPrefix:  theUrl.Path,
			SubPathTag:      strings.Join(theUrl.Query()["tag"], ""),
			BufferedSamples: f.FlagOutputTcpListenBuffer,
		}
		sink.TcpConnLimit = f.FlagTcpConnectionLimit
		if f.FlagTcpLogReceivedData {
			sink.LogReceivedTraffic = log.ErrorLevel
		}
		marshallingSink = &sink.AbstractMarshallingSampleOutput
		resultSink = sink
	default:
		if factory, ok := f.CustomDataSinks[endpoint.Type]; ok && endpoint.IsCustomType {
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
		marshallingSink.Writer = f.Writer()
	}
	return resultSink, nil
}

func (f *EndpointFactory) CreateMarshaller(format MarshallingFormat) (Marshaller, error) {
	factory, ok := f.Marshallers[format]
	if !ok {
		return nil, fmt.Errorf("Unknown marshaller format: %v", format)
	}
	return factory(), nil
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
	Params       map[string]string
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
	case HttpEndpoint:
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

// ParseEndpointDescription parses the given string to an EndpointDescription object.
// The string can be one of two forms: the URL-style description will be parsed by
// ParseUrlEndpointDescription, other descriptions will be parsed by GuessEndpointDescription.
func (f *EndpointFactory) ParseEndpointDescription(endpoint string, isOutput bool) (EndpointDescription, error) {
	if strings.Contains(endpoint, "://") {
		return f.ParseUrlEndpointDescription(endpoint)
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
func (f *EndpointFactory) ParseUrlEndpointDescription(endpoint string) (res EndpointDescription, err error) {
	urlParts := strings.SplitN(endpoint, "://", 2)
	if len(urlParts) != 2 || urlParts[0] == "" || urlParts[1] == "" {
		err = fmt.Errorf("Invalid URL endpoint: %v", endpoint)
		return
	}
	target := urlParts[1]
	res.Target = target
	for _, part := range strings.Split(urlParts[0], "+") {
		// TODO unclean: this parsing method is used for both marshalling/unmarshalling endpoints
		if f.isMarshallingFormat(part) {
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
			case TcpEndpoint, TcpListenEndpoint, FileEndpoint, HttpEndpoint:
				res.Type = EndpointType(part)
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

func (f *EndpointFactory) isMarshallingFormat(formatName string) bool {
	_, ok := f.Marshallers[MarshallingFormat(formatName)]
	return ok
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

type sourceTagger string

func (t sourceTagger) HandleSample(sample *Sample, source string) {
	sample.SetTag(string(t), source)
}
