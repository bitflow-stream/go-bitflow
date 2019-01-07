package bitflow

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SampleTestSuite struct {
	testSuiteBase
}

func TestSamples(t *testing.T) {
	suite.Run(t, new(SampleTestSuite))
}

func (suite *SampleTestSuite) TestEmptySampleRing() {
	s := new(SampleAndHeader)
	ring := NewSampleRing(0)
	suite.True(ring.IsFull())
	suite.Equal(0, ring.Len())
	suite.Empty(ring.Get())
	ring.PushSampleAndHeader(s)
	ring.PushSampleAndHeader(s)
	suite.True(ring.IsFull())
	suite.Equal(0, ring.Len())
	suite.Empty(ring.Get())
}

func (suite *SampleTestSuite) TestSampleRingLen1() {
	s1 := new(SampleAndHeader)
	s2 := new(SampleAndHeader)

	ring := NewSampleRing(1)
	suite.False(ring.IsFull())
	suite.Equal(0, ring.Len())
	suite.Empty(ring.Get())

	ring.PushSampleAndHeader(s1)
	suite.True(ring.IsFull())
	suite.Equal(1, ring.Len())
	suite.Equal([]*SampleAndHeader{s1}, ring.Get())

	ring.PushSampleAndHeader(s2)
	suite.True(ring.IsFull())
	suite.Equal(1, ring.Len())
	suite.Equal([]*SampleAndHeader{s2}, ring.Get())
}

func (suite *SampleTestSuite) TestSampleRingLenN() {
	s1 := new(SampleAndHeader)
	s2 := new(SampleAndHeader)
	s3 := new(SampleAndHeader)
	s4 := new(SampleAndHeader)
	s5 := new(SampleAndHeader)
	s6 := new(SampleAndHeader)
	s7 := new(SampleAndHeader)

	ring := NewSampleRing(3)
	suite.False(ring.IsFull())
	suite.Equal(0, ring.Len())
	suite.Empty(ring.Get())

	ring.PushSampleAndHeader(s1)
	suite.False(ring.IsFull())
	suite.Equal(1, ring.Len())
	suite.Equal([]*SampleAndHeader{s1}, ring.Get())

	ring.PushSampleAndHeader(s2)
	suite.False(ring.IsFull())
	suite.Equal(2, ring.Len())
	suite.Equal([]*SampleAndHeader{s1, s2}, ring.Get())

	ring.PushSampleAndHeader(s3)
	suite.True(ring.IsFull())
	suite.Equal(3, ring.Len())
	suite.Equal([]*SampleAndHeader{s1, s2, s3}, ring.Get())

	ring.PushSampleAndHeader(s4)
	suite.True(ring.IsFull())
	suite.Equal(3, ring.Len())
	suite.Equal([]*SampleAndHeader{s2, s3, s4}, ring.Get())

	ring.PushSampleAndHeader(s5)
	suite.True(ring.IsFull())
	suite.Equal(3, ring.Len())
	suite.Equal([]*SampleAndHeader{s3, s4, s5}, ring.Get())

	ring.PushSampleAndHeader(s6)
	suite.True(ring.IsFull())
	suite.Equal(3, ring.Len())
	suite.Equal([]*SampleAndHeader{s4, s5, s6}, ring.Get())

	ring.PushSampleAndHeader(s7)
	suite.True(ring.IsFull())
	suite.Equal(3, ring.Len())
	suite.Equal([]*SampleAndHeader{s5, s6, s7}, ring.Get())
}
