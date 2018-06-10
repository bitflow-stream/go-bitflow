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

	// TODO rename
	OutgoingSink SampleProcessor
}

// SetSink implements the SampleSource interface.
func (s *AbstractSampleSource) SetSink(sink SampleProcessor) {
	s.OutgoingSink = sink
}

// GetSink implements the SampleSource interface.
func (s *AbstractSampleSource) GetSink() SampleProcessor {
	return s.OutgoingSink
}

// CloseSink closes the subsequent SampleProcessor. It must be called when the
// receiving AbstractSampleSource is stopped. If the wg parameter is not nil,
// The subsequent SampleProcessor is closed in a concurrent goroutine, which is registered
// in the WaitGroup.
func (s *AbstractSampleSource) CloseSink(wg *sync.WaitGroup) {
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
	s.CloseSink(s.wg)
}

// String implements the golib.Task interface.
func (s *EmptySampleSource) String() string {
	return "empty sample source"
}

// SetSampleHandler implements the UnmarshallingSampleSource interface.
func (s *EmptySampleSource) SetSampleHandler(handler ReadSampleHandler) {
	// Do nothing
}
