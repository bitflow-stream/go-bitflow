package sample

import (
	"fmt"
	"io"
	"sync"

	"github.com/antongulenko/golib"
)

// ==================== Data Sink ====================
type MetricSinkBase interface {
	Header(header Header) error
	Sample(sample Sample, header Header) error
}

type MetricSink interface {
	golib.Task
	MetricSinkBase

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
	Writer     SampleWriter
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

// ==================== Empty Source ====================

type EmptyMetricSource struct {
	AbstractMetricSource
	wg *sync.WaitGroup
}

func (s *EmptyMetricSource) Start(wg *sync.WaitGroup) golib.StopChan {
	s.wg = wg
	return nil
}

func (s *EmptyMetricSource) Stop() {
	s.CloseSink(s.wg)
}

func (s *EmptyMetricSource) String() string {
	return "empty metric source"
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

// ==================== Parallel Sample Stream ====================

type ParallelSampleHandler struct {
	BufferedSamples int
	IoBuffer        int
	ParallelParsers int
}

type ParallelSampleStream struct {
	err      error
	incoming chan *BufferedSample
	outgoing chan *BufferedSample
	wg       sync.WaitGroup
	closed   *golib.OneshotCondition
}

func (state *ParallelSampleStream) HasError() bool {
	return state.err != nil && state.err != io.EOF
}

type BufferedSample struct {
	stream   *ParallelSampleStream
	data     []byte
	sample   Sample
	done     bool
	doneCond *sync.Cond
}

func (sample *BufferedSample) WaitDone() error {
	if sample.stream.HasError() {
		return sample.stream.err
	}
	sample.doneCond.L.Lock()
	defer sample.doneCond.L.Unlock()
	for !sample.done && !sample.stream.HasError() {
		sample.doneCond.Wait()
	}
	if sample.stream.HasError() {
		return sample.stream.err
	}
	return nil
}

func (sample *BufferedSample) NotifyDone() {
	sample.doneCond.L.Lock()
	defer sample.doneCond.L.Unlock()
	sample.done = true
	sample.doneCond.Broadcast()
}

// ==================== Synchronizing Metric Sink ====================
// Allows multiple goroutines to write data, synchronizes these writes through a mutex
// Implements MetricSinkBase

type SynchronizingMetricSink struct {
	OutgoingSink MetricSink
	mutex        sync.Mutex
}

func (s *SynchronizingMetricSink) Header(header Header) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.OutgoingSink.Header(header)
}

func (s *SynchronizingMetricSink) Sample(sample Sample, header Header) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.OutgoingSink.Sample(sample, header)
}
