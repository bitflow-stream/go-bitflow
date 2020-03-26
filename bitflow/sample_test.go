package bitflow

import (
	"testing"

	"github.com/antongulenko/golib"
)

type SampleTestSuite struct {
	golib.AbstractTestSuite
}

func TestSamples(t *testing.T) {
	new(SampleTestSuite).Run(t)
}

func (s *SampleTestSuite) TestEmptySampleRing() {
	sample := new(SampleAndHeader)
	ring := NewSampleRing(0)
	s.True(ring.IsFull())
	s.Equal(0, ring.Len())
	s.Empty(ring.Get())
	ring.PushSampleAndHeader(sample)
	ring.PushSampleAndHeader(sample)
	s.True(ring.IsFull())
	s.Equal(0, ring.Len())
	s.Empty(ring.Get())
}

func (s *SampleTestSuite) TestSampleRingLen1() {
	s1 := new(SampleAndHeader)
	s2 := new(SampleAndHeader)

	ring := NewSampleRing(1)
	s.False(ring.IsFull())
	s.Equal(0, ring.Len())
	s.Empty(ring.Get())

	ring.PushSampleAndHeader(s1)
	s.True(ring.IsFull())
	s.Equal(1, ring.Len())
	s.Equal([]*SampleAndHeader{s1}, ring.Get())

	ring.PushSampleAndHeader(s2)
	s.True(ring.IsFull())
	s.Equal(1, ring.Len())
	s.Equal([]*SampleAndHeader{s2}, ring.Get())
}

func (s *SampleTestSuite) TestSampleRingLenN() {
	s1 := new(SampleAndHeader)
	s2 := new(SampleAndHeader)
	s3 := new(SampleAndHeader)
	s4 := new(SampleAndHeader)
	s5 := new(SampleAndHeader)
	s6 := new(SampleAndHeader)
	s7 := new(SampleAndHeader)

	ring := NewSampleRing(3)
	s.False(ring.IsFull())
	s.Equal(0, ring.Len())
	s.Empty(ring.Get())

	ring.PushSampleAndHeader(s1)
	s.False(ring.IsFull())
	s.Equal(1, ring.Len())
	s.Equal([]*SampleAndHeader{s1}, ring.Get())

	ring.PushSampleAndHeader(s2)
	s.False(ring.IsFull())
	s.Equal(2, ring.Len())
	s.Equal([]*SampleAndHeader{s1, s2}, ring.Get())

	ring.PushSampleAndHeader(s3)
	s.True(ring.IsFull())
	s.Equal(3, ring.Len())
	s.Equal([]*SampleAndHeader{s1, s2, s3}, ring.Get())

	ring.PushSampleAndHeader(s4)
	s.True(ring.IsFull())
	s.Equal(3, ring.Len())
	s.Equal([]*SampleAndHeader{s2, s3, s4}, ring.Get())

	ring.PushSampleAndHeader(s5)
	s.True(ring.IsFull())
	s.Equal(3, ring.Len())
	s.Equal([]*SampleAndHeader{s3, s4, s5}, ring.Get())

	ring.PushSampleAndHeader(s6)
	s.True(ring.IsFull())
	s.Equal(3, ring.Len())
	s.Equal([]*SampleAndHeader{s4, s5, s6}, ring.Get())

	ring.PushSampleAndHeader(s7)
	s.True(ring.IsFull())
	s.Equal(3, ring.Len())
	s.Equal([]*SampleAndHeader{s5, s6, s7}, ring.Get())
}
