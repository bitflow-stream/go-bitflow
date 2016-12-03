package bitflow

import (
	"bytes"
	"strconv"
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
		ParallelSampleHandler: ParallelSampleHandler{
			ParallelParsers: 5,
			BufferedSamples: 5,
		},
	}
	stream := writer.Open(&buf, m)
	for i, header := range suite.headers {
		for _, sample := range suite.samples[i] {
			suite.NoError(stream.Sample(sample, header))
		}
	}
	suite.NoError(stream.Close())
	buf.checkClosed()

	// ======== Read ========
	counter := &countingBuf{data: buf.Bytes()}
	sink := testSampleSink{
		suite: suite,
	}
	for i, header := range suite.headers {
		sink.add(suite.samples[i], header)
	}

	source := "Test Source"
	handler := &testSampleHandler{source: source, suite: suite}
	reader := SampleReader{
		ParallelSampleHandler: ParallelSampleHandler{
			ParallelParsers: 5,
			BufferedSamples: 5,
		},
		Handler:      handler,
		Unmarshaller: m,
	}
	readStream := reader.Open(counter, &sink)
	num, err := readStream.ReadSamples(source)
	suite.NoError(err)
	suite.Equal(len(suite.headers)*_samples_per_test, num)

	counter.checkClosed(suite.Assertions)

	suite.Equal(0, len(counter.data))
	sink.checkEmpty()
}

func (suite *TransportStreamTestSuite) testIndividualHeaders(m BidiMarshaller) {
	for i, header := range suite.headers {

		// ======== Write ========
		buf := closingBuffer{
			suite: suite,
		}
		writer := SampleWriter{
			ParallelSampleHandler: ParallelSampleHandler{
				ParallelParsers: 5,
				BufferedSamples: 5,
			},
		}
		stream := writer.Open(&buf, m)
		for _, sample := range suite.samples[i] {
			suite.NoError(stream.Sample(sample, header))
		}
		suite.NoError(stream.Close())
		buf.checkClosed()

		// ======== Read ========
		samples := suite.samples[i]
		counter := &countingBuf{data: buf.Bytes()}
		sink := testSampleSink{
			suite: suite,
		}
		sink.add(samples, header)

		source := "Test Source"
		handler := &testSampleHandler{source: source, suite: suite}
		reader := SampleReader{
			ParallelSampleHandler: ParallelSampleHandler{
				ParallelParsers: 5,
				BufferedSamples: 5,
			},
			Handler:      handler,
			Unmarshaller: m,
		}
		readStream := reader.Open(counter, &sink)
		num, err := readStream.ReadSamples(source)
		suite.NoError(err)
		suite.Equal(len(samples), num)

		counter.checkClosed(suite.Assertions)
		suite.Equal(0, len(counter.data))
		sink.checkEmpty()
	}
}

type testSampleSink struct {
	suite    *TransportStreamTestSuite
	samples  []*Sample
	headers  []*Header
	received int
}

func (s *testSampleSink) add(samples []*Sample, header *Header) {
	s.samples = append(s.samples, samples...)
	for _ = range samples {
		s.headers = append(s.headers, header)
	}
}

func (s *testSampleSink) Sample(sample *Sample, header *Header) error {
	if s.received > len(s.samples)-1 {
		s.suite.Fail("Did not expect any more samples, received " + strconv.Itoa(s.received))
	}
	expectedSample := s.samples[s.received]
	expectedHeader := s.headers[s.received]
	s.received++
	s.suite.compareHeaders(expectedHeader, header)
	s.suite.compareSamples(expectedSample, sample)
	return nil
}

func (s *testSampleSink) checkEmpty() {
	s.suite.Equal(len(s.samples), s.received, "Number of samples received")
}

type testSampleHandler struct {
	suite  *TransportStreamTestSuite
	source string
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

func (h *testSampleHandler) HandleHeader(header *Header, source string) {
	h.suite.Equal(h.source, source)
}

func (h *testSampleHandler) HandleSample(sample *Sample, source string) {
	h.suite.Equal(h.source, source)
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
