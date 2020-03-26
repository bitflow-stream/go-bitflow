package bitflow

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/antongulenko/golib"
)

// Tests for transport_read.go and transport_write.go

type TransportStreamTestSuite struct {
	testSuiteWithSamples
}

func TestTransportStreamTestSuite(t *testing.T) {
	new(TransportStreamTestSuite).Run(t)
}

func (suite *TransportStreamTestSuite) testAllHeaders(m BidiMarshaller) {
	// ======== Write ========
	buf := closingBuffer{
		suite: suite,
	}
	writer := SampleWriter{
		ParallelSampleHandler: parallelHandler,
	}
	stream := writer.Open(&buf, m)
	totalSamples := suite.sendAllSamples(stream)
	suite.NoError(stream.Close())
	buf.checkClosed()

	// ======== Read ========
	counter := &countingBuf{data: buf.Bytes()}
	sink := suite.newFilledTestSink()

	source := "Test Source"
	handler := suite.newHandler(source)
	reader := SampleReader{
		ParallelSampleHandler: parallelHandler,
		Handler:               handler,
		Unmarshaller:          m,
	}
	readStream := reader.Open(counter, sink)
	num, err := readStream.ReadSamples(source)
	suite.NoError(err)
	suite.Equal(totalSamples, num)

	counter.checkClosed(suite.Assertions)

	suite.Equal(0, len(counter.data))
	sink.checkEmpty()
}

func (suite *TransportStreamTestSuite) testIndividualHeaders(m BidiMarshaller) {
	for i := range suite.headers {

		// ======== Write ========
		buf := closingBuffer{
			suite: suite,
		}
		writer := SampleWriter{
			ParallelSampleHandler: parallelHandler,
		}
		stream := writer.Open(&buf, m)
		suite.sendSamples(stream, i)
		suite.NoError(stream.Close())
		buf.checkClosed()

		// ======== Read ========
		samples := suite.samples[i]
		counter := &countingBuf{data: buf.Bytes()}
		sink := suite.newTestSinkFor(i)

		source := "Test Source"
		handler := suite.newHandler(source)
		reader := SampleReader{
			ParallelSampleHandler: parallelHandler,
			Handler:               handler,
			Unmarshaller:          m,
		}
		readStream := reader.Open(counter, sink)
		num, err := readStream.ReadSamples(source)
		suite.NoError(err)
		suite.Equal(len(samples), num)

		counter.checkClosed(suite.Assertions)
		suite.Equal(0, len(counter.data))
		sink.checkEmpty()
	}
}

type closingBuffer struct {
	bytes.Buffer
	closed bool
	suite  *TransportStreamTestSuite
}

func (c *closingBuffer) Close() error {
	c.closed = true
	return nil
}

func (c *closingBuffer) checkClosed() {
	c.suite.True(c.closed, "input stream buffer has not been closed")
}

func (suite *TransportStreamTestSuite) TestTransport_CsvMarshallerSingle() {
	suite.testIndividualHeaders(new(CsvMarshaller))
}

func (suite *TransportStreamTestSuite) TestTransport_CsvMarshallerMulti() {
	suite.testAllHeaders(new(CsvMarshaller))
}

func (suite *TransportStreamTestSuite) TestTransport_BinaryMarshallerSingle() {
	suite.testIndividualHeaders(new(BinaryMarshaller))
}

func (suite *TransportStreamTestSuite) TestTransport_BinaryMarshallerMulti() {
	suite.testAllHeaders(new(BinaryMarshaller))
}

func (suite *TransportStreamTestSuite) TestAllocateSample() {
	var pipe SamplePipeline
	pipe.
		Add(&resizingTestSink{mul: 1, plus: 2}).
		Add(new(NoopProcessor)).
		Add(&resizingTestSink{mul: 1, plus: -100}).
		Add(&resizingTestSink{mul: 3, plus: 1}). // does not change
		Add(&resizingTestSink{mul: 0, plus: 0}). // does not change
		Add(new(NoopProcessor)).
		Add(new(DroppingSampleProcessor))
	pipe.Source = new(EmptySampleSource)
	sink := pipe.Processors[0]

	var group golib.TaskGroup
	pipe.Construct(&group)

	size := RequiredValues(0, sink)
	suite.Equal(7, size)

	size = RequiredValues(1, sink)
	suite.Equal(10, size)

	size = RequiredValues(5, sink)
	suite.Equal(22, size)
}

var _ ResizingSampleProcessor = new(resizingTestSink)

type resizingTestSink struct {
	NoopProcessor
	plus int
	mul  int
}

func (r *resizingTestSink) OutputSampleSize(size int) int {
	return size*r.mul + r.plus
}

func (r *resizingTestSink) String() string {
	return fmt.Sprintf("ResizingTestSink(* %v + %v)", r.mul, r.plus)
}
