package pipeline

import (
	"testing"

	"github.com/antongulenko/go-bitflow"
)

type compare struct {
	t   *testing.T
	win *MetricWindow
	num int
}

func (c *compare) err(msg string, expected interface{}, actual interface{}) {
	c.t.Fatal(c.num, msg, "Expected:", expected, "Actual:", actual, "Window:", c.win)
}

func (c *compare) compare(name string, actual []bitflow.Value, expected []float64) {
	if len(actual) != len(expected) {
		c.err(name+" wrong length", expected, actual)
	}
	for i, expect := range expected {
		if actual[i] != bitflow.Value(expect) {
			c.err(name+" wrong", expected, actual)
		}
	}
}

func (c *compare) do(vals ...float64) {
	empty := len(vals) == 0
	if actual := c.win.Empty(); empty != actual {
		c.err("Wrong Empty()-state", empty, actual)
	}
	if actual := c.win.Size(); len(vals) != actual {
		c.err("Wrong Size()", len(vals), actual)
	}
	winLen := len(c.win.data)
	expectedFull := winLen == len(vals)
	if actual := c.win.Full(); actual != expectedFull {
		c.err("Wrong Full()-state", expectedFull, actual)
	}

	winVals := c.win.Data()
	winFastVals := c.win.FastData()
	c.compare("Data()", winVals, vals)
	c.compare("FastData()", winFastVals, vals)

	filled := make([]bitflow.Value, 3)
	expected := make([]float64, 3)
	copy(expected, vals)
	c.win.FillData(filled)
	c.compare("FillData() [short]", filled, expected)

	filled = make([]bitflow.Value, winLen+5)
	expected = make([]float64, len(filled))
	copy(expected, vals)
	c.win.FillData(filled)
	c.compare("FillData() [long]", filled, expected)

	c.num++
}

func TestWindow(t *testing.T) {
	w := NewMetricWindow(5)
	c := &compare{t: t, win: w}
	c.do()

	w.Push(1)
	c.do(1)
	w.Push(2)
	c.do(1, 2)
	w.Push(3)
	c.do(1, 2, 3)
	w.Push(4)
	c.do(1, 2, 3, 4)
	w.Push(5)
	c.do(1, 2, 3, 4, 5)
	w.Push(6)
	c.do(2, 3, 4, 5, 6)
	w.Push(7)
	c.do(3, 4, 5, 6, 7)
	w.Push(8)
	c.do(4, 5, 6, 7, 8)
	w.Push(9)
	c.do(5, 6, 7, 8, 9)
	w.Push(10)
	c.do(6, 7, 8, 9, 10)
	w.Push(11)
	c.do(7, 8, 9, 10, 11)
	w.Push(12)
	c.do(8, 9, 10, 11, 12)

	w.Pop()
	w.Pop()
	c.do(10, 11, 12)
	w.Pop()
	c.do(11, 12)
	w.Pop()
	w.Push(22)
	c.do(12, 22)
	w.Pop()
	c.do(22)
	w.Pop()
	c.do()
	w.Pop()
	c.do()
	w.Pop()
	c.do()
	w.Push(22)
	c.do(22)
	w.Pop()
	w.Pop()
	c.do()
	w.Pop()
	c.do()
}
