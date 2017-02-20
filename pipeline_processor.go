package bitflow

import (
	"sync"

	"github.com/antongulenko/golib"
)

// SampleProcessor is the combination of MetricSink and MetricSource.
// It receives Samples through the Sample method and
// sends samples to the MetricSink configured over SetSink. The forwarded Samples
// can be the same as received, completely new generated samples, and also a different
// number of Samples from the incoming ones. The Header can also be changed, but then
// the SampleProcessor implementation must take care to adjust the outgoing
// Samples accordingly. All required goroutines must be started in Start()
// and stopped when Close() is called. The Stop() method must be ignored,
// see MetricSink.
type SampleProcessor interface {
	golib.Task
	Sample(sample *Sample, header *Header) error
	SetSink(sink MetricSink)
	Close()
}

// AbstractProcessor is an empty implementation of SampleProcessor. It can be
// directly added to a SamplePipeline and will behave as a no-op processing step.
// Other implementations of SampleProcessor can embed this and override parts of
// the methods as required. No initialization is needed for this type, but an
// instance can only be used once, in one pipeline.
type AbstractProcessor struct {
	AbstractMetricSource
	AbstractMetricSink
	stopChan golib.StopChan
}

// Sample implements the SampleProcessor interface. It performs a sanity check
// by calling p.Check() and then forwards the sample to the configured sink.
func (p *AbstractProcessor) Sample(sample *Sample, header *Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	return p.OutgoingSink.Sample(sample, header)
}

// Check is a utility method that asserts that the receiving AbstractProcessor has
// a sink configured and that the given sample matches the given header. This should
// be done early in every Sample() implementation.
func (p *AbstractProcessor) Check(sample *Sample, header *Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	return nil
}

// Start implements the SampleProcessor interface. It creates an error-channel
// with a small channel buffer. Calling CloseSink() or Error() writes a value
// to that channel to signalize that this AbstractProcessor is finished.
func (p *AbstractProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.stopChan = golib.NewStopChan()
	return p.stopChan
}

// CloseSink reports that this AbstractProcessor is finished processing.
// All goroutines must be stopped, and all Headers and Samples must be already
// forwarded to the outgoing sink, when this is called. CloseSink forwards
// the Close() invokation to the outgoing sink.
func (p *AbstractProcessor) CloseSink() {
	// If there was no error, make sure to signal that this task is done.
	p.stopChan.Stop()
	p.AbstractMetricSource.CloseSink(nil)
}

// Error reports that AbstractProcessor has encountered an error and has stopped
// operation. After calling this, no more Headers and Samples can be forwarded
// to the outgoing sink. Ultimately, p.Close() will be called for cleaning up.
func (p *AbstractProcessor) Error(err error) {
	p.stopChan.StopErr(err)
}

// Close implements the SampleProcessor interface by simply closing the outgoing
// sink. Other types that embed AbstractProcessor can override this to perform
// specific actions when closing.
func (p *AbstractProcessor) Close() {
	// Propagate the Close() invocation
	p.CloseSink()
}

// String implements the SampleProcessor interface. This should be overridden
// by types that are embedding AbstractProcessor.
func (p *AbstractProcessor) String() string {
	return "AbstractProcessor"
}
