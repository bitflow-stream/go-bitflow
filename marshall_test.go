package bitflow

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MarshallerTestSuite struct {
	testSuiteWithSamples
}

func TestMarshallerTestSuite(t *testing.T) {
	suite.Run(t, new(MarshallerTestSuite))
}

func (suite *MarshallerTestSuite) testRead(m BidiMarshaller, rdr *bufio.Reader, expectedHeader *Header, samples []*Sample) {
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

func (suite *MarshallerTestSuite) write(m BidiMarshaller, buf *bytes.Buffer, header *Header, samples []*Sample) {
	suite.NoError(m.WriteHeader(header, buf))
	for _, sample := range samples {
		suite.NoError(m.WriteSample(sample, header, buf))
	}
}

func (suite *MarshallerTestSuite) testAllHeaders(m BidiMarshaller) {
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

func (suite *MarshallerTestSuite) testIndividualHeaders(m BidiMarshaller) {
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

type countingBuf struct {
	data []byte
}

func (c *countingBuf) Read(b []byte) (num int, err error) {
	num = copy(b, c.data)
	if num < len(b) {
		err = io.EOF
	}
	c.data = c.data[num:]
	return
}

func (suite *MarshallerTestSuite) TestCsvMarshallerSingle() {
	suite.testIndividualHeaders(new(CsvMarshaller))
}

func (suite *MarshallerTestSuite) TestCsvMarshallerMulti() {
	suite.testAllHeaders(new(CsvMarshaller))
}

func (suite *MarshallerTestSuite) TestBinaryMarshallerSingle() {
	suite.testIndividualHeaders(new(BinaryMarshaller))
}

func (suite *MarshallerTestSuite) TestBinaryMarshallerMulti() {
	suite.testAllHeaders(new(BinaryMarshaller))
}

type failingBuf struct {
	err error
}

func (c *failingBuf) Read(b []byte) (num int, err error) {
	return 0, c.err
}

func (suite *MarshallerTestSuite) testEOF(m Unmarshaller) {
	buf := failingBuf{err: io.EOF}
	rdr := bufio.NewReader(&buf)
	header, data, err := m.Read(rdr, nil)
	suite.Nil(header)
	suite.Nil(data)
	suite.Equal(io.ErrUnexpectedEOF, err)
}

func (suite *MarshallerTestSuite) TestCsvEOF() {
	suite.testEOF(new(CsvMarshaller))
}

func (suite *MarshallerTestSuite) TestBinaryEOF() {
	suite.testEOF(new(BinaryMarshaller))
}
