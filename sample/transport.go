package sample

import (
	"fmt"
	"io"
	"sync"

	"github.com/antongulenko/golib"
)

// MetricSinkBase is the basic interface to receive/sink samples and headers.
// The main interface for this task is MetricSink, but a few types implement only
// MetricSinkBase, without the additional methods from golib.Task and Close().
type MetricSinkBase interface {
	Header(header *Header) error
	Sample(sample *Sample, header *Header) error
}

// A MetricSink receives samples and headers to do arbitrary operations on them.
// If additional goroutines are required for these operations, they should be created
// after Start() is called. The Stop() method should be ignored, a call to the Close()
// method means that no more header or samples will be coming in, and that all goroutines
// should be stopped. This is to ensure an ordered shutdown of multiple chained
// MetricSinks - the Close() invokation is propagated down the pipeline.
// See the golib.Task interface for info about the Start()/Stop() methods.
type MetricSink interface {
	golib.Task
	MetricSinkBase

	// Should ignore golib.Task.Stop(), but instead close when Close() is called
	// This ensures correct order when shutting down.
	Close()
}

// MarshallingMethicSink extends the MetricSink and allows to set a Marshaller for
// marshalling incoming samples.
type MarshallingMetricSink interface {
	MetricSink
	SetMarshaller(marshaller Marshaller)
}

// AbstractMetricSink is a partial implementation of MetricSink. It simply provides
// an empty implementation of Stop() to emphasize the Stop() should be ignored by
// implementations of MetricSink.
type AbstractMetricSink struct {
}

func (*AbstractMetricSink) Stop() {
	// Should stay empty (implement Close() instead)
}

// AbstractMarshallingMetricSink is a partial implementation of MarshallingMetricSink
// with a simple implementation of SetMarshaller().
type AbstractMarshallingMetricSink struct {
	AbstractMetricSink
	Marshaller Marshaller
	Writer     SampleWriter
}

func (sink *AbstractMarshallingMetricSink) SetMarshaller(marshaller Marshaller) {
	sink.Marshaller = marshaller
}

// MetricSource is the interface used for producing Headers and Samples.
// It should start producing samples in a separate goroutine when Start() is
// called, and should stop all goroutines when Stop() is called. Before Start()
// is called, SetSink() must be called to inform the MetricSource about the MetricSink
// it should output the Headers/Samples into.
// After all samples have been generated (for example because the data source is
// finished, like a file, or because Stop() has been called) the Close() method must be
// called on the configured MetricSink. After calling Close(), no more headers or samples
// are allowed to go into the MetricSink.
// See the golib.Task interface for info about the Start()/Stop() methods.
type MetricSource interface {
	golib.Task
	SetSink(sink MetricSink)
}

// AbstractMetricSource is a partial implementation of MetricSource that stores
// the MetricSink and provides methods to check if the receiving MetricSink
// has been configured, and to close the receiving MetricSink after all samples
// have been generated.
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

// EmptyMetricSource implements MetricSource but does not generate any samples.
// It is used in cases where a source is required but no real implementation is available.
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

// AggregateSink will distribute all incoming Headers and Samples to a slice of
// outgoing MetricSinks. This is used for example to write the same samples both
// to a file and a TCP connection.
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

func (agg AggregateSink) Header(header *Header) error {
	var errors golib.MultiError
	for _, sink := range agg {
		if err := sink.Header(header); err != nil {
			errors.Add(err)
		}
	}
	return errors.NilOrError()
}

func (agg AggregateSink) Sample(sample *Sample, header *Header) error {
	if len(agg) == 0 {
		// Perform sanity check, since no child-sinks are there to perform it
		return sample.Check(header)
	}
	var errors golib.MultiError
	for _, sink := range agg {
		if err := sink.Sample(sample, header); err != nil {
			errors.Add(err)
		}
	}
	return errors.NilOrError()
}

// SynchronizingMetricSink is a MetricSinkBase implementation that allows multiple
// goroutines to write data to the same sink and synchronizes these writes through a mutex.
type SynchronizingMetricSink struct {
	OutgoingSink MetricSink
	mutex        sync.Mutex
}

func (s *SynchronizingMetricSink) Header(header *Header) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.OutgoingSink.Header(header)
}

func (s *SynchronizingMetricSink) Sample(sample *Sample, header *Header) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.OutgoingSink.Sample(sample, header)
}

// ==================== Internal types ====================

type parallelSampleHandler struct {
	BufferedSamples int
	ParallelParsers int
}

type parallelSampleStream struct {
	err    error
	wg     sync.WaitGroup
	closed *golib.OneshotCondition
}

func (state *parallelSampleStream) HasError() bool {
	return state.err != nil && state.err != io.EOF
}

type bufferedSample struct {
	stream   *parallelSampleStream
	data     []byte
	sample   *Sample
	done     bool
	doneCond *sync.Cond
}

func (sample *bufferedSample) WaitDone() {
	sample.doneCond.L.Lock()
	defer sample.doneCond.L.Unlock()
	for !sample.done {
		sample.doneCond.Wait()
	}
}

func (sample *bufferedSample) NotifyDone() {
	sample.doneCond.L.Lock()
	defer sample.doneCond.L.Unlock()
	sample.done = true
	sample.doneCond.Broadcast()
}
