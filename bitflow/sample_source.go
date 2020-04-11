package bitflow

import (
	"sync"

	"github.com/antongulenko/golib"
)

// SampleSource is the interface used for producing Headers and Samples.
// It should start producing samples in a separate goroutine when Start() is
// called, and should stop all goroutines when Close() is called. Before Start()
// is called, SetSink() must be called to inform the SampleSource about the SampleSink
// it should output the Headers/Samples into.
// After all samples have been generated (for example because the data source is
// finished, like a file, or because Close() has been called) the Close() method must be
// called on the outgoing SampleProcessor. After calling Close(), no more headers or samples
// are allowed to go into the SampleProcessor.
// See the golib.Task interface for info about the Start() method.
type SampleSource interface {
	golib.Startable
	String() string
	SetSink(sink SampleProcessor)
	GetSink() SampleProcessor
	Close()
}

// AbstractSampleSource is a partial implementation of SampleSource that stores
// the SampleProcessor and closes the outgoing SampleProcessor after all samples
// have been generated.
type AbstractSampleSource struct {
	out SampleProcessor
}

// SetSink implements the SampleSource interface.
func (s *AbstractSampleSource) SetSink(sink SampleProcessor) {
	s.out = sink
}

// GetSink implements the SampleSource interface.
func (s *AbstractSampleSource) GetSink() SampleProcessor {
	return s.out
}

// CloseSink closes the subsequent SampleProcessor. It must be called after the
// receiving AbstractSampleSource has finished producing samples.
func (s *AbstractSampleSource) CloseSink() {
	if s.out != nil {
		s.out.Close()
	}
}

// CloseSinkParallel closes the subsequent SampleProcessor in a concurrent goroutine, which is registered
// in the WaitGroup. This can be useful compared to CloseSink() in certain cases
// to avoid deadlocks due to long-running Close() invocations. As a general rule of thumb,
// Implementations of SampleSource should use CloseSinkParallel(), while SampleProcessors should simply use CloseSink().
func (s *AbstractSampleSource) CloseSinkParallel(wg *sync.WaitGroup) {
	if s.out != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.out.Close()
		}()
	}
}

// UnmarshallingSampleSource extends SampleSource and adds a configuration setter
// that gives access to the samples that are read by this data source.
type UnmarshallingSampleSource interface {
	SampleSource
	SetSampleHandler(handler ReadSampleHandler)
}

// AbstractUnmarshallingSampleSource extends AbstractSampleSource by adding
// configuration fields required for unmarshalling samples.
type AbstractUnmarshallingSampleSource struct {
	AbstractSampleSource

	// Reader configures aspects of parallel reading and parsing. See SampleReader for more info.
	Reader SampleReader
}

// SetSampleHandler implements the UnmarshallingSampleSource interface
func (s *AbstractUnmarshallingSampleSource) SetSampleHandler(handler ReadSampleHandler) {
	s.Reader.Handler = handler
}

// EmptySampleSource implements SampleSource but does not generate any samples.
// It is used in cases where a source is required but no real implementation is available.
type EmptySampleSource struct {
	AbstractSampleSource
	wg *sync.WaitGroup
}

// Start implements the golib.Task interface.
func (s *EmptySampleSource) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	s.wg = wg
	return
}

// Close implements the SampleSource interface.
func (s *EmptySampleSource) Close() {
	s.CloseSinkParallel(s.wg)
}

// String implements the golib.Task interface.
func (s *EmptySampleSource) String() string {
	return "empty sample source"
}

// SetSampleHandler implements the UnmarshallingSampleSource interface.
func (s *EmptySampleSource) SetSampleHandler(_ ReadSampleHandler) {
	// Do nothing
}

// ClosedSampleSource implements SampleSource, but instead generating samples, it immediately closes.
// It can be used for debugging purposes, to create a real pipeline that does not
type ClosedSampleSource struct {
	AbstractSampleSource
	wg *sync.WaitGroup
}

// Start implements the golib.Task interface.
func (s *ClosedSampleSource) Start(wg *sync.WaitGroup) golib.StopChan {
	s.wg = wg
	return golib.NewStoppedChan(nil)
}

// Close implements the SampleSource interface.
func (s *ClosedSampleSource) Close() {
	s.CloseSinkParallel(s.wg)
}

// String implements the golib.Task interface.
func (s *ClosedSampleSource) String() string {
	return "closed sample source"
}

// SetSampleHandler implements the UnmarshallingSampleSource interface.
func (s *ClosedSampleSource) SetSampleHandler(_ ReadSampleHandler) {
	// Do nothing
}
