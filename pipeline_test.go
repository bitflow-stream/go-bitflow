package bitflow

import (
	"testing"

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

	// Test StdEndpoint specially
	compare("std://-", UndefinedFormat, TextFormat, StdEndpoint, "-")
	compare("std+csv://-", CsvFormat, CsvFormat, StdEndpoint, "-")
	compare("csv+std://-", CsvFormat, CsvFormat, StdEndpoint, "-")
	compare("std+bin://-", BinaryFormat, BinaryFormat, StdEndpoint, "-")
	compare("bin+std://-", BinaryFormat, BinaryFormat, StdEndpoint, "-")
	compare("std+text://-", TextFormat, TextFormat, StdEndpoint, "-")
	compare("text+std://-", TextFormat, TextFormat, StdEndpoint, "-")
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
}
