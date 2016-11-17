package pipeline

import "github.com/antongulenko/go-bitflow"

type MetricWindow struct {
	data  []bitflow.Value
	first int
	next  int
	full  bool
}

func NewMetricWindow(size int) *MetricWindow {
	if size == 0 { // Empty window is not valid
		size = 1
	}
	return &MetricWindow{
		data: make([]bitflow.Value, size),
	}
}

func (w *MetricWindow) inc(i int) int {
	return (i + 1) % len(w.data)
}

func (w *MetricWindow) Push(val bitflow.Value) {
	w.data[w.next] = val
	w.next = w.inc(w.next)
	if w.full {
		w.first = w.inc(w.first)
	} else if w.first == w.next {
		w.full = true
	}
}

func (w *MetricWindow) Size() int {
	switch {
	case w.full:
		return len(w.data)
	case w.first <= w.next:
		return w.next - w.first
	default: // w.first > w.next
		return len(w.data) - w.first + w.next
	}
}

func (w *MetricWindow) Empty() bool {
	return w.first == w.next && !w.full
}

func (w *MetricWindow) Full() bool {
	return w.full
}

// Remove and return the oldest value. The oldest value is also deleted by
// Push() when the window is full.
func (w *MetricWindow) Pop() bitflow.Value {
	if w.Empty() {
		return 0
	}
	val := w.data[w.first]
	w.first = w.inc(w.first)
	w.full = false
	return val
}

// Avoid copying if possible. Dangerous.
func (w *MetricWindow) FastData() []bitflow.Value {
	if w.Empty() {
		return nil
	}
	if w.first < w.next {
		return w.data[w.first:w.next]
	} else {
		return w.Data()
	}
}

func (w *MetricWindow) Data() []bitflow.Value {
	res := make([]bitflow.Value, len(w.data))
	res = w.FillData(res)
	return res
}

func (w *MetricWindow) FillData(target []bitflow.Value) []bitflow.Value {
	if w.Empty() {
		return nil
	}
	var length int
	if w.first < w.next {
		length = copy(target, w.data[w.first:w.next])
	} else {
		length = copy(target, w.data[w.first:])
		if len(target) > length {
			length += copy(target[length:], w.data[:w.next])
		}
	}
	return target[:length]
}
