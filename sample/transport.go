package sample

import (
	"fmt"
	"sync"

	"github.com/antongulenko/golib"
)

// ==================== Data Sink ====================
type MetricSink interface {
	golib.Task
	Header(header Header) error
	Sample(sample Sample, header Header) error

	// Should ignore golib.Task.Stop(), but instead close when Close() is called
	// This ensures correct order when shutting down.
	Close()
}

type MarshallingMetricSink interface {
	MetricSink
	SetMarshaller(marshaller Marshaller)
}

type AbstractMetricSink struct {
}

func (*AbstractMetricSink) Stop() {
	// Should stay empty (implement Close() instead)
}

type AbstractMarshallingMetricSink struct {
	AbstractMetricSink
	Marshaller Marshaller
}

func (sink *AbstractMarshallingMetricSink) SetMarshaller(marshaller Marshaller) {
	sink.Marshaller = marshaller
}

// ==================== Data Source ====================
type MetricSource interface {
	golib.Task
	SetSink(sink MetricSink)
}

type AbstractMetricSource struct {
	OutgoingSink MetricSink
}

func (s *AbstractMetricSource) SetSink(sink MetricSink) {
	s.OutgoingSink = sink
}

func (s *AbstractMetricSource) CheckSink() error {
	if s.OutgoingSink == nil {
		return fmt.Errorf("No data sink set for %v", s)
	}
	return nil
}

func (s *AbstractMetricSource) CloseSink(wg *sync.WaitGroup) {
	// Must be called when this source is stopped
	if s.OutgoingSink != nil {
		if wg == nil {
			s.OutgoingSink.Close()
		} else {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s.OutgoingSink.Close()
			}()
		}
	}
}

type UnmarshallingMetricSource interface {
	MetricSource
	SetUnmarshaller(unmarshaller Unmarshaller) // Must be called before Start()
}

type AbstractUnmarshallingMetricSource struct {
	AbstractMetricSource
	Unmarshaller Unmarshaller
}

func (s *AbstractUnmarshallingMetricSource) SetUnmarshaller(unmarshaller Unmarshaller) {
	s.Unmarshaller = unmarshaller
}

// ==================== Aggregating Sink ====================
type AggregateSink []MetricSink

func (agg AggregateSink) String() string {
	return fmt.Sprintf("AggregateSink(len %v)", len(agg))
}

// The golib.Task interface cannot really be supported here
func (agg AggregateSink) Start(wg *sync.WaitGroup) golib.StopChan {
	panic("Start should not be called on AggregateSink")
}

func (agg AggregateSink) Stop() {
	panic("Stop should not be called on AggregateSink")
}

func (agg AggregateSink) Close() {
	for _, sink := range agg {
		sink.Close()
	}
}

func (agg AggregateSink) SetMarshaller(marshaller Marshaller) {
	for _, sink := range agg {
		if um, ok := sink.(MarshallingMetricSink); ok {
			um.SetMarshaller(marshaller)
		}
	}
}

func (agg AggregateSink) Header(header Header) error {
	var errors golib.MultiError
	for _, sink := range agg {
		if err := sink.Header(header); err != nil {
			errors.Add(err)
		}
	}
	return errors.NilOrError()
}

func (agg AggregateSink) Sample(sample Sample, header Header) error {
	var errors golib.MultiError
	for _, sink := range agg {
		if err := sink.Sample(sample, header); err != nil {
			errors.Add(err)
		}
	}
	return errors.NilOrError()
}
