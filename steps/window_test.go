package steps

import (
	"testing"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/stretchr/testify/suite"
)

type WindowTestSuite struct {
	golib.AbstractTestSuite
}

func TestWindow(t *testing.T) {
	suite.Run(t, new(WindowTestSuite))
}

func (s *WindowTestSuite) do(win *MetricWindow, floatValues ...float64) {
	var values []bitflow.Value
	if floatValues != nil {
		values = make([]bitflow.Value, len(floatValues))
		for i, val := range floatValues {
			values[i] = bitflow.Value(val)
		}
	}

	s.Equal(len(values) == 0, win.Empty())
	s.Equal(len(values), win.Size())
	winLen := len(win.data)
	s.Equal(winLen == len(values), win.Full())

	s.Equal(values, win.Data(), "Data()")
	s.Equal(values, win.FastData(), "FastData()")

	filled := make([]bitflow.Value, 3)
	expected := make([]bitflow.Value, 3)
	copy(expected, values)
	win.FillData(filled)
	s.Equal(expected, filled, "FillData() [short]")

	filled = make([]bitflow.Value, winLen+5)
	expected = make([]bitflow.Value, len(filled))
	copy(expected, values)
	win.FillData(filled)
	s.Equal(expected, filled, "FillData() [long]")
}

func (s *WindowTestSuite) TestEmptyWindow() {
	w := NewMetricWindow(0) // Should behave like window size 1
	s.do(w)

	w.Push(1)
	s.do(w, 1)
	w.Push(2)
	w.Push(2)
	s.do(w, 2)
	w.Pop()
	w.Pop()
	s.do(w)
}

func (s *WindowTestSuite) TestWindow() {
	w := NewMetricWindow(5)
	s.do(w)

	w.Push(1)
	s.do(w, 1)
	w.Push(2)
	s.do(w, 1, 2)
	w.Push(3)
	s.do(w, 1, 2, 3)
	w.Push(4)
	s.do(w, 1, 2, 3, 4)
	w.Push(5)
	s.do(w, 1, 2, 3, 4, 5)
	w.Push(6)
	s.do(w, 2, 3, 4, 5, 6)
	w.Push(7)
	s.do(w, 3, 4, 5, 6, 7)
	w.Push(8)
	s.do(w, 4, 5, 6, 7, 8)
	w.Push(9)
	s.do(w, 5, 6, 7, 8, 9)
	w.Push(10)
	s.do(w, 6, 7, 8, 9, 10)
	w.Push(11)
	s.do(w, 7, 8, 9, 10, 11)
	w.Push(12)
	s.do(w, 8, 9, 10, 11, 12)

	w.Pop()
	w.Pop()
	s.do(w, 10, 11, 12)
	w.Pop()
	s.do(w, 11, 12)
	w.Pop()
	w.Push(22)
	s.do(w, 12, 22)
	w.Pop()
	s.do(w, 22)
	w.Pop()
	s.do(w)
	w.Pop()
	s.do(w)
	w.Pop()
	s.do(w)
	w.Push(22)
	s.do(w, 22)
	w.Pop()
	w.Pop()
	s.do(w)
	w.Pop()
	s.do(w)
}
