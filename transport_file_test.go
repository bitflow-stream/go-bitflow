package bitflow

import (
	"io/ioutil"
	"strings"
	"sync"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type FileTestSuite struct {
	testSuiteWithSamples
}

func TestFileTransport(t *testing.T) {
	suite.Run(t, new(FileTestSuite))
}

func (suite *FileTestSuite) testAllHeaders(m Marshaller) {
	basename := "bitflow-test-file"
	newName := "TEST-123123"
	testFileFile, err := ioutil.TempFile("", basename+"."+m.String()+".")
	testFile := testFileFile.Name()
	suite.NoError(err)
	log.Debugln("TEST FILE for", m, ":", testFile)
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
	out.Stop()
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
			suite.True(strings.Contains(name, basename))
			return newName
		},
	}
	in.Reader.ParallelSampleHandler = parallel_handler
	in.Reader.Handler = suite.newHandler(newName)
	in.SetSink(testSink)
	ch = in.Start(&wg)
	wg.Wait()
	in.Stop()
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
		basename := "bitflow-test-file"
		newName := "TEST-123123"
		testFileFile, err := ioutil.TempFile("", basename+"."+m.String()+".")
		testFile := testFileFile.Name()
		testFiles = append(testFiles, testFile)
		suite.NoError(err)
		log.Debugln("TEST FILE for", m, ":", testFile)

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
		out.Stop()
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
				suite.True(strings.Contains(name, basename))
				return newName
			},
		}
		in.Reader.ParallelSampleHandler = parallel_handler
		in.Reader.Handler = suite.newHandler(newName)
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
