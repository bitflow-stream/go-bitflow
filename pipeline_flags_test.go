package bitflow

import (
	"fmt"
	"testing"
	"time"

	"github.com/antongulenko/golib/gotermBox"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type PipelineTestSuite struct {
	t *testing.T
	*require.Assertions
}

func (suite *PipelineTestSuite) T() *testing.T {
	return suite.t
}

func (suite *PipelineTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func TestPipelineTestSuite(t *testing.T) {
	suite.Run(t, new(PipelineTestSuite))
}

func (suite *PipelineTestSuite) TestGuessEndpoint() {
	compare := func(endpoint string, format MarshallingFormat, typ EndpointType) {
		desc, err := ParseEndpointDescription(endpoint)
		suite.NoError(err)
		suite.Equal(EndpointDescription{Format: UndefinedFormat, Type: typ, Target: endpoint}, desc)
		suite.Equal(format, desc.OutputFormat())
	}
	compareErr2 := func(endpoint string, errStr string) {
		desc, err := ParseEndpointDescription(endpoint)
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

	compare("-", TextFormat, StdEndpoint)

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
	compare("host:8888/xxx", BinaryFormat, TcpEndpoint)

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
	compare := func(endpoint string, format MarshallingFormat, outputFormat MarshallingFormat, typ EndpointType, target string) {
		desc, err := ParseEndpointDescription(endpoint)
		suite.NoError(err)
		suite.Equal(EndpointDescription{Format: format, Type: typ, Target: target}, desc)
		suite.Equal(outputFormat, desc.OutputFormat())
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

	// Test ConsoleBoxEndpoint (no variations)
	compare("box://-", UndefinedFormat, UndefinedFormat, ConsoleBoxEndpoint, "-")
}

func (suite *PipelineTestSuite) TestUrlEndpointErrors() {
	err := func(endpoint string, errStr string) {
		_, err := ParseEndpointDescription(endpoint)
		suite.Error(err)
		suite.Contains(err.Error(), errStr)
	}

	err("://", "Invalid URL")
	err("://xxx", "Invalid URL")
	err("file://", "Invalid URL")

	err("x://xxx", "Illegal transport")
	err("x+x+x://xxx", "Illegal transport")
	err("csv+xx://x", "Illegal transport")

	err("csv+csv://x", "Multiple formats")
	err("bin+bin://x", "Multiple formats")
	err("text+text://x", "Multiple formats")
	err("bin+csv+xx://x", "Multiple formats")
	err("csv+text+text://x", "Multiple formats")

	err("file+file://x", "Multiple transport")
	err("tcp+tcp://x", "Multiple transport")
	err("std+std://-", "Multiple transport")
	err("csv+file+file://x", "Multiple transport")
	err("csv+file+x://x", "Multiple transport")
	err("file+csv+file://x", "Multiple transport")

	err("std://x", "Transport 'std' can only be defined with target '-'")
	err("csv+std://x", "Transport 'std' can only be defined with target '-'")
	err("std+csv://x", "Transport 'std' can only be defined with target '-'")

	err("box://x", "Transport 'box' can only be defined with target '-'")
	err("box+csv://x", "Transport 'box' can only be defined with target '-'")
	err("csv+box://-", "Cannot define the format for transport 'box'")
	err("box+box://-", "Multiple transport")
}

func (suite *PipelineTestSuite) make_factory() EndpointFactory {
	return EndpointFactory{
		testmode:                  true,
		FlagInputFilesRobust:      true,
		FlagOutputFilesClean:      true,
		FlagIoBuffer:              666,
		FlagOutputTcpListenBuffer: 777,
		FlagTcpConnectionLimit:    10,
		FlagInputTcpAcceptLimit:   20,
		FlagTcpDropErrors:         true,
		FlagParallelHandler:       parallel_handler,
	}
}

func (suite *PipelineTestSuite) Test_no_inputs() {
	factory := suite.make_factory()
	source, err := factory.CreateInput()
	suite.NoError(err)
	suite.Equal(nil, source)
}

func (suite *PipelineTestSuite) Test_input_file() {
	factory := suite.make_factory()
	files := []string{"file1", "file2", "file3"}
	handler := &testSampleHandler{source: "xxx"}
	source, err := factory.CreateInput(files...)
	source.SetSampleHandler(handler)
	suite.NoError(err)
	expected := &FileSource{
		Filenames: files,
		Robust:    true,
		IoBuffer:  666,
	}
	expected.Reader.Handler = handler
	expected.Reader.ParallelSampleHandler = parallel_handler
	suite.Equal(expected, source)
}

func (suite *PipelineTestSuite) Test_input_tcp() {
	factory := suite.make_factory()
	hosts := []string{"host1:123", "host2:2", "host2:5"}
	handler := &testSampleHandler{source: "xxx"}
	source, err := factory.CreateInput(hosts...)
	source.SetSampleHandler(handler)
	suite.NoError(err)
	expected := &TCPSource{
		RemoteAddrs:   hosts,
		PrintErrors:   false,
		RetryInterval: tcp_download_retry_interval,
		DialTimeout:   tcp_dial_timeout,
	}
	expected.TcpConnLimit = 10
	expected.Reader.Handler = handler
	expected.Reader.ParallelSampleHandler = parallel_handler
	suite.Equal(expected, source)
}

func (suite *PipelineTestSuite) Test_input_tcp_listen() {
	factory := suite.make_factory()
	endpoint := ":123"
	handler := &testSampleHandler{source: "xxx"}
	source, err := factory.CreateInput(endpoint)
	source.SetSampleHandler(handler)
	suite.NoError(err)
	expected := NewTcpListenerSource(endpoint)
	expected.SimultaneousConnections = 20
	expected.TcpConnLimit = 10
	expected.Reader.Handler = handler
	expected.Reader.ParallelSampleHandler = parallel_handler
	suite.Equal(expected, source)
}

func (suite *PipelineTestSuite) Test_input_std() {
	factory := suite.make_factory()
	endpoint := "-"
	handler := &testSampleHandler{source: "xxx"}
	source, err := factory.CreateInput(endpoint)
	source.SetSampleHandler(handler)
	suite.NoError(err)
	expected := NewConsoleSource()
	expected.Reader.Handler = handler
	expected.Reader.ParallelSampleHandler = parallel_handler
	suite.Equal(expected, source)
}

func (suite *PipelineTestSuite) Test_input_multiple() {
	test := func(input1, input2 string, inputs ...string) {
		factory := suite.make_factory()
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
	factory := suite.make_factory()

	source, err := factory.CreateInput("abc://x")
	suite.Error(err, "Unknown endpoint input type 'abc'")
	suite.Nil(source)

	source, err = factory.CreateInput("box://x")
	suite.Error(err, "Unknown endpoint input type 'box'")
	suite.Nil(source)

	sink, err := factory.CreateOutput("abc://x")
	suite.Error(err, "Unknown endpoint input type 'abc'")
	suite.Nil(sink)
}

func (suite *PipelineTestSuite) Test_input_multiple_listener() {
	factory := suite.make_factory()
	source, err := factory.CreateInput(":123", ":456")
	suite.Error(err)
	suite.Equal(err.Error(), fmt.Sprintf("Cannot listen for input on multiple TCP ports"))
	suite.Nil(source)
}

func (suite *PipelineTestSuite) Test_input_multiple_std() {
	factory := suite.make_factory()
	source, err := factory.CreateInput("-", "-")
	suite.Error(err)
	suite.Equal(err.Error(), fmt.Sprintf("Cannot read from stdin multiple times"))
	suite.Nil(source)
}

func (suite *PipelineTestSuite) Test_outputs() {
	test := func(output string, expected MetricSink) {
		factory := suite.make_factory()
		sink, err := factory.CreateOutput(output)
		suite.NoError(err)
		suite.Equal(expected, sink)
	}

	setup := func(sink *AbstractMarshallingMetricSink, format string) {
		sink.Marshaller = MarshallingFormat(format).Marshaller()
		sink.Writer.ParallelSampleHandler = parallel_handler
	}

	box_settings := gotermBox.CliLogBox{
		NoUtf8:        true,
		LogLines:      22,
		MessageBuffer: 1000,
	}
	box_interval := 666 * time.Millisecond

	// Bad test: sets global configuration state
	ConsoleBoxSettings = box_settings
	ConsoleBoxUpdateInterval = box_interval

	box := func() MetricSink {
		s := &ConsoleBoxSink{
			CliLogBox:      box_settings,
			UpdateInterval: box_interval,
		}
		return s
	}

	std := func(format string) MetricSink {
		s := NewConsoleSink()
		setup(&s.AbstractMarshallingMetricSink, format)
		return s
	}
	file := func(filename string, format string) MetricSink {
		s := &FileSink{
			Filename:   filename,
			IoBuffer:   666,
			CleanFiles: true,
		}
		setup(&s.AbstractMarshallingMetricSink, format)
		return s
	}
	tcp := func(endpoint string, format string) MetricSink {
		s := &TCPSink{
			Endpoint:    endpoint,
			PrintErrors: false,
			DialTimeout: tcp_dial_timeout,
		}
		s.TcpConnLimit = 10
		setup(&s.AbstractMarshallingMetricSink, format)
		return s
	}
	listen := func(endpoint string, format string) MetricSink {
		s := &TCPListenerSink{
			Endpoint:        endpoint,
			BufferedSamples: 777,
		}
		s.TcpConnLimit = 10
		setup(&s.AbstractMarshallingMetricSink, format)
		return s
	}

	// Individual outputs
	test("box://-", box())
	test("-", std("text"))
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
