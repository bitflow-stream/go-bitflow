package bitflow

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
	Sample(sample *Sample, header *Header) error
}

// HeaderChecker is a helper type for implementations of MetricSinkBase
// to find out, when the incoming header changes.
type HeaderChecker struct {
	LastHeader *Header
}

// HeaderChanged returns true, if the newHeader parameter represents a different header
// from the last time HeaderChanged was called. The result will also be true for
// the first time this method is called.
func (h *HeaderChecker) HeaderChanged(newHeader *Header) bool {
	changed := !newHeader.Equals(h.LastHeader)
	h.LastHeader = newHeader
	return changed
}

// InitializedHeaderChanged returns true, if the newHeader parameter represents a different header
// from the last time HeaderChanged was called. The first call to this method
// will return false, so this can be used in situations where the header has to be initialized.
func (h *HeaderChecker) InitializedHeaderChanged(newHeader *Header) bool {
	if h.LastHeader == nil {
		h.LastHeader = newHeader
		return false
	}
	changed := !newHeader.Equals(h.LastHeader)
	h.LastHeader = newHeader
	return changed
}

// A MetricSink receives samples and headers to do arbitrary operations on them.
// If additional goroutines are required for these operations, they should be created
// after Start() is called. The Stop() method should be ignored, a call to the Close()
// method means that no more header or samples will be coming in, and that all goroutines
// should be stopped. This is to ensure an ordered shutdown of multiple chained
// MetricSinks - the Close() invocation is propagated down the pipeline.
// See the golib.Task interface for info about the Start()/Stop() methods.
type MetricSink interface {
	golib.Task
	MetricSinkBase

	// Should ignore golib.Task.Stop(), but instead close when Close() is called
	// This ensures correct order when shutting down.
	Close()
}

// MarshallingMetricSink extends the MetricSink and allows to set a Marshaller for
// marshalling incoming samples.
type MarshallingMetricSink interface {
	MetricSink
	SetMarshaller(marshaller Marshaller)
}

// AbstractMetricSink is a partial implementation of MetricSink. It simply provides
// an empty implementation of Stop() to emphasize that Stop() should be ignored by
// implementations of MetricSink.
type AbstractMetricSink struct {
}

// Stop implements the golib.Task interface.
func (*AbstractMetricSink) Stop() {
	// Should stay empty (implement Close() instead)
}

// AbstractMarshallingMetricSink is a partial implementation of MarshallingMetricSink
// with a simple implementation of SetMarshaller().
type AbstractMarshallingMetricSink struct {
	AbstractMetricSink

	// Marshaller will be used when converting Samples to byte buffers before
	// writing them to the given output stream.
	Marshaller Marshaller

	// Writer contains variables that controll the marshalling and writing process.
	// They must be configured before calling Start() on this AbstractMarshallingMetricSink.
	Writer SampleWriter
}

// SetMarshaller implements the MarshallingMetricSink interface.
func (sink *AbstractMarshallingMetricSink) SetMarshaller(marshaller Marshaller) {
	sink.Marshaller = marshaller
}

// ResizingMetricSink is a helper interface that can be implemented by SampleProcessors
// in order to make AbstractSampleSource.AllocateSample() more reliable. The result of
// the OutputSampleSize() method should give a worst-case estimation of the number of values
// that will be present in Samples after this SampleProcessor is done processing a sample.
// This allows the optimization of pre-allocating a value array large enough to hold the final
// amount of metrics.
// The optimization works best when all samples are processed in a one-to-one fashion,
// i.e. no samples are split into multiple samples.
type ResizingMetricSink interface {
	MetricSink
	OutputSampleSize(sampleSize int) int
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
	GetSink() MetricSink
}

// AbstractMetricSource is a partial implementation of MetricSource that stores
// the MetricSink and provides methods to check if the receiving MetricSink
// has been configured, and to close the receiving MetricSink after all samples
// have been generated.
type AbstractMetricSource struct {
	OutgoingSink MetricSink
}

// SetSink implements the MetricSource interface.
func (s *AbstractMetricSource) SetSink(sink MetricSink) {
	s.OutgoingSink = sink
}

// GetSink implements the MetricSource interface.
func (s *AbstractMetricSource) GetSink() MetricSink {
	return s.OutgoingSink
}

// CheckSink is a helper method that returns an error if the SetSink() has
// not been called on the receiving AbstractMetricSource.
func (s *AbstractMetricSource) CheckSink() error {
	if s.OutgoingSink == nil {
		return fmt.Errorf("No data sink set for %v", s)
	}
	return nil
}

// CloseSink closes the outgoing MetricSink. It must be called when the
// receiving AbstractMetricSource is stopped. If the wg parameter is not nil,
// The outgoing MetricSink is closed in a concurrent goroutine, which is registered
// in the WaitGroup.
func (s *AbstractMetricSource) CloseSink(wg *sync.WaitGroup) {
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

// RequiredValues the number of Values that should be large enough to hold
// the end-result after processing a Sample by all intermediate SampleProcessors.
// The result is based on ResizingMetricSink.OutputSampleSize(). MetricSink instances
// that do not implement the ResizingMetricSink interface are assumed to not increase the
// number metrics.
func RequiredValues(numFields int, sink MetricSinkBase) int {
	for {
		if sink == nil {
			break
		}
		if sink, ok := sink.(ResizingMetricSink); ok {
			newSize := sink.OutputSampleSize(numFields)
			if newSize > numFields {
				numFields = newSize
			}
		}
		if source, ok := sink.(MetricSource); ok {
			sink = source.GetSink()
		} else {
			break
		}
	}
	return numFields
}

// UnmarshallingMetricSource extends MetricSource and adds a configuration setter
// that gives access to the samples that are read by this data source.
type UnmarshallingMetricSource interface {
	MetricSource
	SetSampleHandler(handler ReadSampleHandler)
}

// AbstractUnmarshallingMetricSource extends AbstractMetricSource by adding
// configuration fields required for unmarshalling samples.
type AbstractUnmarshallingMetricSource struct {
	AbstractMetricSource

	// Reader configures aspects of parallel reading and parsing. See SampleReader for more info.
	Reader SampleReader
}

// SetSampleHandler implements the UnmarshallingMetricSource interface
func (s *AbstractUnmarshallingMetricSource) SetSampleHandler(handler ReadSampleHandler) {
	s.Reader.Handler = handler
}

// EmptyMetricSource implements MetricSource but does not generate any samples.
// It is used in cases where a source is required but no real implementation is available.
type EmptyMetricSource struct {
	AbstractMetricSource
	wg *sync.WaitGroup
}

// Start implements the golib.Task interface.
func (s *EmptyMetricSource) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	s.wg = wg
	return
}

// Stop implements the golib.Task interface.
func (s *EmptyMetricSource) Stop() {
	s.CloseSink(s.wg)
}

// String implements the golib.Task interface.
func (s *EmptyMetricSource) String() string {
	return "empty metric source"
}

// Stop implements the UnmarshallingMetricSource interface.
func (s *EmptyMetricSource) SetSampleHandler(handler ReadSampleHandler) {
	// Do nothing
}

// EmptyMetricSink implements the MetricSink interface without outputting the
// received samples anywhere.
type EmptyMetricSink struct {
	AbstractMetricSink
}

// Start implements the golib.Task interface.
func (s *EmptyMetricSink) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	return
}

// String implements the MetricSink interface.
func (s *EmptyMetricSink) Close() {
}

// String implements the golib.Task interface.
func (s *EmptyMetricSink) String() string {
	return "empty metric sink"
}

// Sample implements the MetricSink interface.
func (s *EmptyMetricSink) Sample(sample *Sample, header *Header) error {
	return nil
}

// SynchronizingMetricSink is a MetricSinkBase implementation that allows multiple
// goroutines to write data to the same sink and synchronizes these writes through a mutex.
type SynchronizingMetricSink struct {
	OutgoingSink MetricSink
	mutex        sync.Mutex
}

// Sample implements the MetricSinkBase interface.
func (s *SynchronizingMetricSink) Sample(sample *Sample, header *Header) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.OutgoingSink.Sample(sample, header)
}

// ==================== Configuration types ====================

// ParallelSampleHandler is a configuration type that is included in
// SampleReader and SampleWriter. Both the reader and writer can marshall
// and unmarshall Samples in parallel, and these routines are controlled
// through the two parameters in ParallelSampleHandler.
type ParallelSampleHandler struct {
	// BufferedSamples is the number of Samples that are buffered between the
	// marshall/unmarshall routines and the routine that writes/reads the input
	// or output streams.
	// The purpose of the buffer is, for example, to allow the routine reading a file
	// to read the data for multiple Samples in one read operation, which then
	// allows the parallel parsing routines to parse all the read Samples at the same time.
	// Setting BufferedSamples is a trade-off between memory consumption and
	// parallelism, but most of the time a value of around 1000 or so should be enough.
	// If this value is not set, no parallelism will be possible because
	// the channel between the cooperating routines will block on each operation.
	BufferedSamples int

	// ParallelParsers can be set to the number of goroutines that will be
	// used when marshalling or unmarshalling samples. These routines can
	// parallelize the parsing and marshalling operations. The most benefit
	// from the parallelism comes when reading samples from e.g. files, because
	// reading the file into memory can be decoupled from parsing Samples,
	// and multiple Samples can be parsed at the same time.
	//
	// This must be set to a value greater than zero, otherwise no goroutines
	// will be started.
	ParallelParsers int
}

// ==================== Internal types ====================

type parallelSampleStream struct {
	err     golib.MultiError
	errLock sync.Mutex

	wg     sync.WaitGroup
	closed golib.StopChan
}

func (state *parallelSampleStream) addError(err error) bool {
	if err != nil {
		state.errLock.Lock()
		defer state.errLock.Unlock()
		state.err.Add(err)
		return true
	}
	return false
}

func (state *parallelSampleStream) hasError() bool {
	state.errLock.Lock()
	defer state.errLock.Unlock()
	if len(state.err) > 0 {
		for _, err := range state.err {
			if err != io.EOF {
				return true
			}
		}
	}
	return false
}

func (state *parallelSampleStream) getErrorNoEOF() error {
	state.errLock.Lock()
	defer state.errLock.Unlock()
	var result golib.MultiError
	if len(state.err) > 0 {
		for _, err := range state.err {
			if err != io.EOF {
				result.Add(err)
			}
		}
	}
	return result.NilOrError()
}

type bufferedSample struct {
	stream   *parallelSampleStream
	data     []byte
	header   *Header // Used for marshalling and unmarshalling/parsing
	sample   *Sample
	done     bool
	doneCond *sync.Cond
}

func (sample *bufferedSample) waitDone() {
	sample.doneCond.L.Lock()
	defer sample.doneCond.L.Unlock()
	for !sample.done {
		sample.doneCond.Wait()
	}
}

func (sample *bufferedSample) notifyDone() {
	sample.doneCond.L.Lock()
	defer sample.doneCond.L.Unlock()
	sample.done = true
	sample.doneCond.Broadcast()
}
