package bitflow

import (
	"sync"

	"github.com/antongulenko/golib"
)

// SampleProcessor is the combination of MetricSink and MetricSource.
// It receives Headers and Samples through the Header and Sample methods and
// sends samples to the MetricSink configured over SetSink. The forwarded Samples
// can be the same as received, completely new generated samples, and also a different
// number of Samples from the incoming ones. The Header can also be changed, but hten
// the SampleProcessor implementation must take care to adjust the outgoing
// Samples accordingly. All required goroutines must be started in Start()
// and stopped when Close() is called. The Stop() method must be ignored,
// see MetricSink.
type SampleProcessor interface {
	golib.Task
	Header(header *Header) error
	Sample(sample *Sample, header *Header) error
	SetSink(sink MetricSink)
	Close()
}

// SamplePipeline reads data from a source, pipes it through zero or more SampleProcessor
// instances and outputs the final Samples into a MetricSink.
// The job of the SamplePipeline is to connect all the processing steps
// in the Construct method. After calling Construct, the SamplePipeline should not
// used any further.
type SamplePipeline struct {
	Source     MetricSource
	Processors []SampleProcessor
	Sink       MetricSink
}

// Construct connects the MetricSource, all SampleProcessors and the MetricSink
// inside the receiving SamplePipeline. It adds all these golib.Task instances
// to the golib.TaskGroup parameter. Afterwards, tasks.WaitAndStop() cann be called
// to start the entire pipeline. At least the Source and the Sink field
// must be set in the pipeline. The Construct method will not fail without them,
// but starting the resulting pipeline will not work.
func (p *SamplePipeline) Construct(tasks *golib.TaskGroup) {
	// First connect all sources with their sinks
	source := p.Source
	for _, processor := range p.Processors {
		if source != nil {
			source.SetSink(processor)
		}
		source = processor
	}
	if source != nil {
		source.SetSink(p.Sink)
	}

	// Then add all tasks in reverse: start the final sink first.
	// Each sink must be started before the source can push data into it.
	if agg, ok := p.Sink.(AggregateSink); ok {
		for _, sink := range agg {
			tasks.Add(sink)
		}
	} else {
		tasks.Add(p.Sink)
	}
	for i := len(p.Processors) - 1; i >= 0; i-- {
		tasks.Add(p.Processors[i])
	}
	tasks.Add(p.Source)
}

// Add adds the SampleProcessor parameter to the list of SampleProcessors in the
// receiving SamplePipeline. The Source and Sink fields must be accessed directly.
// The Processors field can also be accessed directly, but the Add method allows
// chaining multiple Add invokations like so:
//   pipeline.Add(processor1).Add(processor2)
func (p *SamplePipeline) Add(processor SampleProcessor) *SamplePipeline {
	if processor != nil {
		p.Processors = append(p.Processors, processor)
	}
	return p
}

// AbstractProcessor is an empty implementation of SampleProcessor. It can be
// directly added to a SamplePipeline and will behave as a no-op processing step.
// Other implementations of SampleProcessor can embed this and override parts of
// the methods as required. No initialization is needed for this type, but an
// instance can only be used once, in one pipeline.
type AbstractProcessor struct {
	AbstractMetricSource
	AbstractMetricSink
	stopChan chan error
}

// Header implements the SampleProcessor interface. It checks if a sink has been
// defined for this processor and then forwards the header to that sink.
func (p *AbstractProcessor) Header(header *Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		return p.OutgoingSink.Header(header)
	}
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
	// Chan buffer of 2 makes sure the Processor can report an error and then
	// call AbstractProcessor.CloseSink() without worrying if the error has
	// been reported previously or not.
	p.stopChan = make(chan error, 2)
	return p.stopChan
}

// CloseSink reports that this AbstractProcessor is finished processing.
// All goroutines must be stopped, and all Headers and Samples must be already
// forwarded to the outgoing sink, when this is called. CloseSink forwards
// the Close() invokation to the outgoing sink.
func (p *AbstractProcessor) CloseSink(wg *sync.WaitGroup) {
	if c := p.stopChan; c != nil {
		// If there was no error, make sure the channel still returns something to signal that this task is done.
		c <- nil
		p.stopChan = nil
	}
	p.AbstractMetricSource.CloseSink(wg)
}

// Error reports that AbstractProcessor has encountered an error and has stopped
// operation. After calling this, no more Headers and Samples can be forwarded
// to the outgoing sink. Ultimately, p.Close() will be called for cleaning up.
func (p *AbstractProcessor) Error(err error) {
	if c := p.stopChan; c != nil {
		c <- err
	}
}

// Close implements the SampleProcessor interface by simply closing the outgoing
// sink. Other types that embed AbstractProcessor can override this to perform
// specific actions when closing.
func (p *AbstractProcessor) Close() {
	// Propagate the Close() invocation
	p.CloseSink(nil)
}

// String implements the SampleProcessor interface. This should be overridden
// by types that are embedding AbstractProcessor.
func (p *AbstractProcessor) String() string {
	return "AbstractProcessor"
}
