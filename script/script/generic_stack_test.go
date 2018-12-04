package script

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var sampleProp0 = SampleType{id: "0"}
var sampleProp1 = SampleType{id: "1"}

type SampleType struct {
	id string
}

func TestPushPeekSingle(t *testing.T) {
	var s GenericStack
	s.Push(sampleProp0)
	s.Push(sampleProp1)
	prop := s.PeekSingle()
	prop = s.PeekSingle()
	prop = s.PeekSingle()
	assert.Equal(t, prop, sampleProp1)
	assert.Equal(t, 2, len(s))
}

func TestPushPopSingle(t *testing.T) {
	var s GenericStack
	s.Push(sampleProp0)
	s.Push(sampleProp1)
	assert.Equal(t, s.PopSingle(), sampleProp1)
	assert.Equal(t, s.PeekSingle(), sampleProp0)
	assert.Equal(t, 1, len(s))
}

func TestPushMulti(t *testing.T) {
	var s GenericStack
	s.Push(sampleProp0, sampleProp1)
	assert.Equal(t, s.Peek(), []interface{}{sampleProp0, sampleProp1})
	assert.Equal(t, s.Pop(), []interface{}{sampleProp0, sampleProp1})
	assert.Equal(t, 0, len(s))
}

func TestPopSinglePanic(t *testing.T) {
	var s GenericStack
	s.Push(sampleProp0, sampleProp1)
	assert.Panics(t, func() {
		s.PeekSingle()
	})

	s = GenericStack{}
	s.Push(sampleProp0, sampleProp1)
	assert.Panics(t, func() {
		s.PopSingle()
	})

	s = GenericStack{}
	s.Push(sampleProp0, sampleProp1)
	s.Push(sampleProp0)
	s.PopSingle()
	assert.Panics(t, func() {
		s.PopSingle()
	})
}

func TestPanicEmpty(t *testing.T) {
	assert.Panics(t, func() {
		var s GenericStack
		s.PopSingle()
	})
	assert.Panics(t, func() {
		var s GenericStack
		s.PeekSingle()
	})
	assert.Panics(t, func() {
		var s GenericStack
		s.Push(sampleProp1)
		s.Pop()
		s.PeekSingle()
	})
}

func TestSampleTypeStackWith5Elements_whenLen_thenReturn5(t *testing.T) {
	var s GenericStack
	s.Push(sampleProp0, sampleProp1, sampleProp0)
	s.Push(sampleProp0)
	s.Push(sampleProp0, sampleProp0)
	s.Push(sampleProp0)
	s.Push(sampleProp0)
	assert.Equal(t, 5, len(s))
	s.Pop()
	s.Pop()
	assert.Equal(t, 3, len(s))
}
