package sample

import "github.com/antongulenko/golib"

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
