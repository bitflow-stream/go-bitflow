package bitflow

import (
	"errors"
	"fmt"
	"testing"

	"github.com/antongulenko/golib"
	"github.com/stretchr/testify/suite"
)

type PipelineTestSuite struct {
	golib.AbstractTestSuite
}

func TestPipelineTestSuite(t *testing.T) {
	suite.Run(t, new(PipelineTestSuite))
}

func (suite *PipelineTestSuite) TestGuessEndpoint() {
	factory := NewEndpointFactory()
	compareX := func(endpoint string, format MarshallingFormat, typ EndpointType, isCustom bool, isOutput bool) {
		desc, err := factory.ParseEndpointDescription(endpoint, isOutput)
		suite.NoError(err)
		suite.Equal(EndpointDescription{Format: UndefinedFormat, Type: typ, Target: endpoint, IsCustomType: isCustom}, desc)
		suite.Equal(format, desc.OutputFormat())
	}
	compare := func(endpoint string, format MarshallingFormat, typ EndpointType) {
		compareX(endpoint, format, typ, false, false)
	}
	compareErr2 := func(endpoint string, errStr string) {
		desc, err := factory.ParseEndpointDescription(endpoint, false)
		suite.Error(err)
		suite.Contains(err.Error(), errStr)
		suite.Equal(EndpointDescription{Format: UndefinedFormat, Type: UndefinedEndpoint, Target: endpoint}, desc)
		suite.Panics(func() {
			desc.OutputFormat()
		})
	}
	compareErr := func(endpoint string) {
		compareErr2(endpoint, "Not a filename and not a valid TCP endpoint")
	}

	compareX("-", TextFormat, StdEndpoint, false, true)
	compareX("-", TextFormat, StdEndpoint, false, false)

	// File names
	compare("xxx", CsvFormat, FileEndpoint)
	compare("xxx.csv", CsvFormat, FileEndpoint)
	compare("xxx.xxx.xxx", CsvFormat, FileEndpoint)
	compare("xxx.bin", BinaryFormat, FileEndpoint)

	// TCP endpoints
	compare(":8888", BinaryFormat, TcpListenEndpoint)
	compare("localhost:8888", BinaryFormat, TcpEndpoint)
	compare("a.b.c:8888", BinaryFormat, TcpEndpoint)
	compare("192.168.0.0:8888", BinaryFormat, TcpEndpoint)

	// Neither file names nor valid TCP endpoints
	compareErr(":ABC")
	compareErr(":88888888888")
	compareErr(":ABC/abc")
	compareErr(":8888:xxx")
	compareErr(":8888:xxx.bin")
	compareErr("host:ABC")
	compareErr("host:88888888888")
	compareErr("host:ABC/abc")
	compareErr("host:8888:xxx")
	compareErr("host:8888:xxx.bin")
	compareErr2("", "Empty endpoint/file is not valid")
}

func (suite *PipelineTestSuite) TestUrlEndpoint() {
	factory := NewEndpointFactory()
	compare := func(endpoint string, format MarshallingFormat, outputFormat MarshallingFormat, typ EndpointType, target string) {
		// Check for output
		desc, err := factory.ParseEndpointDescription(endpoint, true)
		suite.NoError(err)
		suite.Equal(EndpointDescription{Format: format, Type: typ, Target: target, IsCustomType: false}, desc)
		suite.Equal(outputFormat, desc.OutputFormat())

		// Check for input
		if format != TextFormat {
			desc, err := factory.ParseEndpointDescription(endpoint, false)
			suite.NoError(err)
			suite.Equal(EndpointDescription{Format: format, Type: typ, Target: target, IsCustomType: false}, desc)
			suite.Equal(outputFormat, desc.OutputFormat())
		}
	}

	checkOne := func(typ EndpointType, outputFormat MarshallingFormat) {
		s := string(typ)
		compare(s+":///xxx", UndefinedFormat, outputFormat, typ, "/xxx")
		compare(s+":///x/xx", UndefinedFormat, outputFormat, typ, "/x/xx")
		compare(s+"://:9999", UndefinedFormat, outputFormat, typ, ":9999")
		compare(s+"://localhost:9999", UndefinedFormat, outputFormat, typ, "localhost:9999")
		compare(s+"://localhost:9999/xxx/xxx", UndefinedFormat, outputFormat, typ, "localhost:9999/xxx/xxx")

		compare(s+"://xxx", UndefinedFormat, outputFormat, typ, "xxx")
		compare(s+"://xxx.csv", UndefinedFormat, outputFormat, typ, "xxx.csv")
		compare(s+"://xxx.xxx", UndefinedFormat, outputFormat, typ, "xxx.xxx")

		if typ == FileEndpoint {
			outputFormat = BinaryFormat
		}
		compare(s+"://xxx.bin", UndefinedFormat, outputFormat, typ, "xxx.bin")
	}
	checkMix := func(typ EndpointType) {
		s := string(typ)
		compare(s+"+csv://xxx.csv", CsvFormat, CsvFormat, typ, "xxx.csv")
		compare("csv+"+s+"://xxx.bin", CsvFormat, CsvFormat, typ, "xxx.bin")
		compare(s+"+bin://xxx.csv", BinaryFormat, BinaryFormat, typ, "xxx.csv")
		compare("bin+"+s+"://xxx.bin", BinaryFormat, BinaryFormat, typ, "xxx.bin")
		compare(s+"+text://xxx.csv", TextFormat, TextFormat, typ, "xxx.csv")
		compare("text+"+s+"://xxx.bin", TextFormat, TextFormat, typ, "xxx.bin")
	}
	checkFormat := func(format MarshallingFormat) {
		s := string(format) + "://"
		compare(s+"-", format, format, StdEndpoint, "-")
		compare(s+"xxx", format, format, FileEndpoint, "xxx")
		compare(s+"xxx.bin", format, format, FileEndpoint, "xxx.bin")
		compare(s+":8888", format, format, TcpListenEndpoint, ":8888")
		compare(s+"localhost:8888", format, format, TcpEndpoint, "localhost:8888")
	}

	checkOne(FileEndpoint, CsvFormat)
	checkOne(TcpEndpoint, BinaryFormat)
	checkOne(TcpListenEndpoint, BinaryFormat)

	checkMix(FileEndpoint)
	checkMix(TcpEndpoint)
	checkMix(TcpListenEndpoint)

	checkFormat(BinaryFormat)
	checkFormat(CsvFormat)
	checkFormat(TextFormat)

	// Test StdEndpoint
	compare("std://-", UndefinedFormat, TextFormat, StdEndpoint, "-")
	compare("std+csv://-", CsvFormat, CsvFormat, StdEndpoint, "-")
	compare("csv+std://-", CsvFormat, CsvFormat, StdEndpoint, "-")
	compare("std+bin://-", BinaryFormat, BinaryFormat, StdEndpoint, "-")
	compare("bin+std://-", BinaryFormat, BinaryFormat, StdEndpoint, "-")
	compare("std+text://-", TextFormat, TextFormat, StdEndpoint, "-")
	compare("text+std://-", TextFormat, TextFormat, StdEndpoint, "-")
}

func (suite *PipelineTestSuite) TestUrlEndpointErrors() {
	factory := NewEndpointFactory()
	err := func(endpoint string, errStr string) {
		_, err := factory.ParseEndpointDescription(endpoint, true)
		suite.Error(err)
		suite.Contains(err.Error(), errStr)
	}

	err("://", "Invalid URL")
	err("://xxx", "Invalid URL")
	err("file://", "Invalid URL")

	err("csv+csv://x", "Multiple formats")
	err("bin+bin://x", "Multiple formats")
	err("text+text://x", "Multiple formats")
	err("bin+csv+xx://x", "Multiple formats")
	err("csv+text+text://x", "Multiple formats")

	err("xx+xx://x", "Multiple transport")
	err("file+file://x", "Multiple transport")
	err("tcp+tcp://x", "Multiple transport")
	err("std+std://-", "Multiple transport")
	err("csv+file+file://x", "Multiple transport")
	err("csv+file+x://x", "Multiple transport")
	err("file+csv+file://x", "Multiple transport")

	err("std://x", "Transport 'std' can only be defined with target '-'")
	err("csv+std://x", "Transport 'std' can only be defined with target '-'")
	err("std+csv://x", "Transport 'std' can only be defined with target '-'")

	err("csv+box://-", "Cannot define the data format for transport 'box'")
	err("box+box://-", "Multiple transport")
}

func (suite *PipelineTestSuite) makeFactory() *EndpointFactory {
	f := NewEndpointFactory()
	f.FlagInputFilesRobust = true
	f.FlagOutputFilesClean = true
	f.FlagIoBuffer = 666
	f.FlagOutputTcpListenBuffer = 777
	f.FlagTcpConnectionLimit = 10
	f.FlagInputTcpAcceptLimit = 20
	f.FlagTcpSourceDropErrors = true
	f.FlagParallelHandler = parallelHandler
	return f
}

func (suite *PipelineTestSuite) Test_no_inputs() {
	factory := suite.makeFactory()
	source, err := factory.CreateInput()
	suite.NoError(err)
	suite.Equal(nil, source)
}

func (suite *PipelineTestSuite) Test_input_file() {
	factory := suite.makeFactory()
	files := []string{"file1", "file2", "file3"}
	handler := &testSampleHandler{source: "xxx"}
	source, err := factory.CreateInput(files...)
	suite.NoError(err)
	source.(UnmarshallingSampleSource).SetSampleHandler(handler)
	expected := &FileSource{
		FileNames: files,
		Robust:    true,
		IoBuffer:  666,
	}
	expected.Reader.Handler = handler
	expected.Reader.ParallelSampleHandler = parallelHandler
	suite.Equal(expected, source)
}

func (suite *PipelineTestSuite) Test_input_tcp() {
	factory := suite.makeFactory()
	hosts := []string{"host1:123", "host2:2", "host2:5"}
	handler := &testSampleHandler{source: "xxx"}
	source, err := factory.CreateInput(hosts...)
	suite.NoError(err)
	source.(UnmarshallingSampleSource).SetSampleHandler(handler)
	expected := &TCPSource{
		RemoteAddrs:   hosts,
		PrintErrors:   false,
		RetryInterval: tcpDownloadRetryInterval,
		DialTimeout:   tcpDialTimeout,
	}
	expected.TcpConnLimit = 10
	expected.Reader.Handler = handler
	expected.Reader.ParallelSampleHandler = parallelHandler
	suite.Equal(expected, source)
}

func (suite *PipelineTestSuite) Test_input_tcp_listen() {
	factory := suite.makeFactory()
	endpoint := ":123"
	handler := &testSampleHandler{source: "xxx"}
	source, err := factory.CreateInput(endpoint)
	suite.NoError(err)
	source.(UnmarshallingSampleSource).SetSampleHandler(handler)
	expected := NewTcpListenerSource(endpoint)
	expected.SimultaneousConnections = 20
	expected.TcpConnLimit = 10
	expected.Reader.Handler = handler
	expected.Reader.ParallelSampleHandler = parallelHandler
	suite.Equal(expected, source)
}

func (suite *PipelineTestSuite) Test_input_std() {
	factory := suite.makeFactory()
	endpoint := "-"
	handler := &testSampleHandler{source: "xxx"}
	source, err := factory.CreateInput(endpoint)
	suite.NoError(err)
	source.(UnmarshallingSampleSource).SetSampleHandler(handler)
	expected := NewConsoleSource()
	expected.Reader.Handler = handler
	expected.Reader.ParallelSampleHandler = parallelHandler
	suite.Equal(expected, source)
}

func (suite *PipelineTestSuite) Test_input_multiple() {
	test := func(input1, input2 string, inputs ...string) {
		factory := suite.makeFactory()
		source, err := factory.CreateInput(inputs...)
		suite.Error(err)
		suite.Equal(err.Error(), fmt.Sprintf("Please provide only one data source (Provided %s and %s)", input1, input2))
		suite.Nil(source)
	}

	test("tcp", "file", "host:123", "file1", "file2")
	test("file", "tcp", "fileA", "fileB", "host:123", "file1", "file2")
	test("tcp", "listen", "host:123", ":123", "file1")
	test("listen", "std", ":123", "-", "host:123")
	test("std", "listen", "-", ":123", "-", "host:123")
}

func (suite *PipelineTestSuite) Test_unknown_endpoint_type() {
	factory := suite.makeFactory()

	source, err := factory.CreateInput("abc://x")
	suite.EqualError(err, "Unknown input endpoint type: abc")
	suite.Nil(source)

	source, err = factory.CreateInput("box://x")
	suite.EqualError(err, "Unknown input endpoint type: box")
	suite.Nil(source)

	sink, err := factory.CreateOutput("abc://x")
	suite.EqualError(err, "Unknown output endpoint type: abc")
	suite.Nil(sink)
}

func (suite *PipelineTestSuite) Test_input_multiple_listener() {
	factory := suite.makeFactory()
	source, err := factory.CreateInput(":123", ":456")
	suite.Error(err)
	suite.Equal(err.Error(), fmt.Sprint("Cannot listen for input on multiple TCP ports"))
	suite.Nil(source)
}

func (suite *PipelineTestSuite) Test_input_multiple_std() {
	factory := suite.makeFactory()
	source, err := factory.CreateInput("-", "-")
	suite.Error(err)
	suite.Equal(err.Error(), fmt.Sprint("Cannot read from stdin multiple times"))
	suite.Nil(source)
}

func (suite *PipelineTestSuite) Test_outputs() {
	test := func(output string, expected SampleSink) {
		factory := suite.makeFactory()
		sink, err := factory.CreateOutput(output)
		suite.NoError(err)
		suite.Equal(expected, sink)
	}

	setup := func(sink *AbstractMarshallingSampleOutput, format string, isConsole bool) {
		switch format {
		case "bin":
			sink.Marshaller = BinaryMarshaller{}
		case "csv":
			sink.Marshaller = &CsvMarshaller{}
		case "text":
			sink.Marshaller = TextMarshaller{AssumeStdout: isConsole}
		default:
			panic("Illegal format: " + format)
		}
		sink.Writer.ParallelSampleHandler = parallelHandler
	}

	std := func(format string) SampleSink {
		s := NewConsoleSink()
		setup(&s.AbstractMarshallingSampleOutput, format, true)
		return s
	}
	file := func(filename string, format string) SampleSink {
		s := &FileSink{
			Filename:   filename,
			IoBuffer:   666,
			CleanFiles: true,
		}
		setup(&s.AbstractMarshallingSampleOutput, format, false)
		return s
	}
	tcp := func(endpoint string, format string) SampleSink {
		s := &TCPSink{
			Endpoint:    endpoint,
			DialTimeout: tcpDialTimeout,
		}
		s.TcpConnLimit = 10
		setup(&s.AbstractMarshallingSampleOutput, format, false)
		return s
	}
	listen := func(endpoint string, format string) SampleSink {
		s := &TCPListenerSink{
			Endpoint:        endpoint,
			BufferedSamples: 777,
		}
		s.TcpConnLimit = 10
		setup(&s.AbstractMarshallingSampleOutput, format, false)
		return s
	}

	// Individual outputs
	test("-", std("text"))
	test("std://-", std("text"))
	test("text+std://-", std("text"))
	test("std+csv://-", std("csv"))
	test("bin+std://-", std("bin"))
	test("csv://-", std("csv"))
	test("bin://-", std("bin"))
	test("text://-", std("text"))

	test("file", file("file", "csv"))
	test("csv://file", file("file", "csv"))
	test("bin://file", file("file", "bin"))
	test("text://file", file("file", "text"))

	test("host:123", tcp("host:123", "bin"))
	test("csv://host:123", tcp("host:123", "csv"))
	test("bin://host:123", tcp("host:123", "bin"))
	test("text://host:123", tcp("host:123", "text"))

	test(":123", listen(":123", "bin"))
	test("csv://:123", listen(":123", "csv"))
	test("bin://:123", listen(":123", "bin"))
	test("text://:123", listen(":123", "text"))
}

func (suite *PipelineTestSuite) Test_custom_endpoints() {
	factory := suite.makeFactory()
	testEndpointType := EndpointType("testendpoint")
	testSource := NewConsoleSource()
	testSink := NewConsoleSink()
	var expectedTarget string
	var injectedError error
	factory.CustomDataSources[testEndpointType] = func(target string) (SampleSource, error) {
		suite.Equal(expectedTarget, target)
		return testSource, injectedError
	}
	factory.CustomDataSinks[testEndpointType] = func(target string) (SampleProcessor, error) {
		suite.Equal(expectedTarget, target)
		return testSink, injectedError
	}

	var res interface{}
	var err error

	expectedTarget = "xxx"
	res, err = factory.CreateInput("testendpoint://xxx", "testendpoint://yyy")
	suite.EqualError(err, "Cannot define multiple sources for custom input type 'testendpoint'")
	suite.Nil(res)

	expectedTarget = "xxx"
	res, err = factory.CreateInput("testendpoint://xxx")
	suite.NoError(err)
	suite.Equal(res, testSource)

	expectedTarget = "yyy"
	res, err = factory.CreateOutput("testendpoint://yyy")
	suite.NoError(err)
	suite.Equal(res, testSink)

	injectedError = errors.New("TEST-ERROR")

	expectedTarget = "xxx"
	res, err = factory.CreateInput("testendpoint://xxx")
	suite.EqualError(err, "Error creating 'testendpoint' input: TEST-ERROR")
	suite.Equal(res, nil)

	expectedTarget = "yyy"
	res, err = factory.CreateOutput("testendpoint://yyy")
	suite.EqualError(err, "Error creating 'testendpoint' output: TEST-ERROR")
	suite.Equal(res, nil)
}
