package bitflow

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Tests for transport_read.go and transport_write.go

type TransportStreamTestSuite struct {
	testSuiteWithSamples
}

func TestTransportStreamTestSuite(t *testing.T) {
	suite.Run(t, new(TransportStreamTestSuite))
}

func (suite *TransportStreamTestSuite) testAllHeaders(m BidiMarshaller) {
	// ======== Write ========
	buf := closingBuffer{
		suite: suite,
	}
	writer := SampleWriter{
		ParallelSampleHandler: parallel_handler,
	}
	stream := writer.Open(&buf, m)
	total_samples := suite.sendAllSamples(stream)
	suite.NoError(stream.Close())
	buf.checkClosed()

	// ======== Read ========
	counter := &countingBuf{data: buf.Bytes()}
	sink := suite.newFilledTestSink()

	source := "Test Source"
	handler := suite.newHandler(source)
	reader := SampleReader{
		ParallelSampleHandler: parallel_handler,
		Handler:               handler,
		Unmarshaller:          m,
	}
	readStream := reader.Open(counter, sink)
	num, err := readStream.ReadSamples(source)
	suite.NoError(err)
	suite.Equal(total_samples, num)

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
			ParallelSampleHandler: parallel_handler,
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
			ParallelSampleHandler: parallel_handler,
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