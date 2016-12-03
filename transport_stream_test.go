package bitflow

import (
	"bufio"
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

func (suite *TransportStreamTestSuite) testRead(m BidiMarshaller, rdr *bufio.Reader, expectedHeader *Header, samples []*Sample) {
	header, data, err := m.Read(rdr, nil)
	suite.NoError(err)
	suite.Nil(data)
	suite.NotNil(header)
	suite.Equal(expectedHeader.HasTags, header.HasTags, "Header.HasTags")
	suite.Equal(expectedHeader.Fields, header.Fields, "Header.Fields")

	for _, expectedSample := range samples {
		nilHeader, data, err := m.Read(rdr, header)
		suite.NoError(err)
		suite.NotNil(data)
		suite.Nil(nilHeader)
		sample, err := m.ParseSample(header, data)
		suite.NoError(err)
		suite.compareSamples(expectedSample, sample)
	}
}

func (suite *TransportStreamTestSuite) write(m BidiMarshaller, buf *bytes.Buffer, header *Header, samples []*Sample) {
	suite.NoError(m.WriteHeader(header, buf))
	for _, sample := range samples {
		suite.NoError(m.WriteSample(sample, header, buf))
	}
}

func (suite *TransportStreamTestSuite) testAllHeaders(m BidiMarshaller) {
	var buf bytes.Buffer
	for i, header := range suite.headers {
		suite.write(m, &buf, header, suite.samples[i])
	}

	counter := &countingBuf{data: buf.Bytes()}
	rdr := bufio.NewReader(counter)
	for i, header := range suite.headers {
		suite.testRead(m, rdr, header, suite.samples[i])
	}
	suite.Equal(0, len(counter.data))
	_, err := rdr.ReadByte()
	suite.Error(err)
}

func (suite *TransportStreamTestSuite) testIndividualHeaders(m BidiMarshaller) {
	for i, header := range suite.headers {
		var buf bytes.Buffer
		suite.write(m, &buf, header, suite.samples[i])

		counter := &countingBuf{data: buf.Bytes()}
		rdr := bufio.NewReader(counter)
		suite.testRead(m, rdr, header, suite.samples[i])
		suite.Equal(0, len(counter.data))
		_, err := rdr.ReadByte()
		suite.Error(err)
	}
}

func (suite *TransportStreamTestSuite) TestCsvMarshallerSingle() {
	suite.testIndividualHeaders(new(CsvMarshaller))
}

func (suite *TransportStreamTestSuite) TestCsvMarshallerMulti() {
	suite.testAllHeaders(new(CsvMarshaller))
}

func (suite *TransportStreamTestSuite) TestBinaryMarshallerSingle() {
	suite.testIndividualHeaders(new(BinaryMarshaller))
}

func (suite *TransportStreamTestSuite) TestBinaryMarshallerMulti() {
	suite.testAllHeaders(new(BinaryMarshaller))
}
