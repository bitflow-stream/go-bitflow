package bitflow

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type FileTestSuite struct {
	testSuiteWithSamples

	dir       string
	fileIndex int
}

const (
	baseFilename        = "bitflow-test-file"
	replacementFilename = "TEST-123123"
)

func TestFileTransport(t *testing.T) {
	suite.Run(t, new(FileTestSuite))
}

func (suite *FileTestSuite) SetupSuite() {
	suite.testSuiteWithSamples.SetupTest()
	dir, err := ioutil.TempDir("", "tests")
	suite.NoError(err)
	suite.dir = dir
}

func (suite *FileTestSuite) TearDownSuite() {
	suite.NoError(os.RemoveAll(suite.dir))
}

func (suite *FileTestSuite) getTestFile(m Marshaller) string {
	suite.fileIndex++
	testFile := path.Join(suite.dir, fmt.Sprintf("%v-%v.%v", baseFilename, suite.fileIndex, m.String()))
	log.Debugln("TEST FILE for", m, ":", testFile)
	return testFile
}

func (suite *FileTestSuite) testAllHeaders(m Marshaller) {
	testFile := suite.getTestFile(m)
	defer func() {
		g := NewFileGroup(testFile)
		suite.NoError(g.DeleteFiles())
	}()

	// ========= Write file
	out := &FileSink{
		Filename:   testFile,
		IoBuffer:   1024,
		CleanFiles: true,
	}
	out.SetMarshaller(m)
	out.Writer.ParallelSampleHandler = parallel_handler
	var wg sync.WaitGroup
	ch := out.Start(&wg)
	suite.sendAllSamples(out)
	out.Close()
	wg.Wait()
	ch.Wait()
	suite.NoError(ch.Err())

	// ========= Read file
	testSink := suite.newFilledTestSink()
	in := &FileSource{
		FileNames:      []string{testFile},
		ReadFileGroups: true,
		Robust:         false,
		IoBuffer:       1024,
		ConvertFilename: func(name string) string {
			suite.True(strings.Contains(name, baseFilename))
			return replacementFilename
		},
	}
	in.Reader.ParallelSampleHandler = parallel_handler
	in.Reader.Handler = suite.newHandler(replacementFilename)
	in.SetSink(testSink)
	ch = in.Start(&wg)
	wg.Wait()
	in.Close()
	ch.Wait()
	suite.NoError(ch.Err())
	testSink.checkEmpty()
}

func (suite *FileTestSuite) testIndividualHeaders(m Marshaller) {
	var testFiles []string
	defer func() {
		for _, testFile := range testFiles {
			g := NewFileGroup(testFile)
			suite.NoError(g.DeleteFiles())
		}
	}()
	for i := range suite.headers {
		testFile := suite.getTestFile(m)
		testFiles = append(testFiles, testFile)

		// ========= Write file
		out := &FileSink{
			Filename:   testFile,
			IoBuffer:   1024,
			CleanFiles: true,
		}
		out.SetMarshaller(m)
		out.Writer.ParallelSampleHandler = parallel_handler
		var wg sync.WaitGroup
		ch := out.Start(&wg)
		suite.sendSamples(out, i)
		out.Close()
		wg.Wait()
		ch.Wait()
		suite.NoError(ch.Err())

		// ========= Read file
		testSink := suite.newTestSinkFor(i)
		in := &FileSource{
			FileNames: []string{testFile},
			Robust:    false,
			IoBuffer:  1024,
			ConvertFilename: func(name string) string {
				suite.True(strings.Contains(name, baseFilename))
				return replacementFilename
			},
		}
		in.Reader.ParallelSampleHandler = parallel_handler
		in.Reader.Handler = suite.newHandler(replacementFilename)
		in.SetSink(testSink)
		ch = in.Start(&wg)
		wg.Wait()
		suite.NoError(ch.Err())
		testSink.checkEmpty()
	}
}

func (suite *FileTestSuite) TestFilesIndividualCsv() {
	suite.testIndividualHeaders(new(CsvMarshaller))
}

func (suite *FileTestSuite) TestFilesIndividualBinary() {
	suite.testIndividualHeaders(new(BinaryMarshaller))
}

func (suite *FileTestSuite) TestFilesAllCsv() {
	suite.testAllHeaders(new(CsvMarshaller))
}

func (suite *FileTestSuite) TestFilesAllBinary() {
	suite.testAllHeaders(new(BinaryMarshaller))
}
