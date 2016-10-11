package sample

import "github.com/antongulenko/golib"

// Combination of MetricSink and MetricSource
type SampleProcessor interface {
	golib.Task
	Header(header *Header) error
	Sample(sample *Sample, header *Header) error
	SetSink(sink MetricSink)
	Close()
}

// A pipeline reads data from a source, pipes it through some processing steps
// and outputs it into a final sink.
type SamplePipeline struct {
	Source     MetricSource
	Processors []SampleProcessor
	Sink       MetricSink
}

// After this, tasks.WaitAndExit() or similar can be called to run the pipeline
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

func (p *SamplePipeline) Add(processor SampleProcessor) *SamplePipeline {
	if processor != nil {
		p.Processors = append(p.Processors, processor)
	}
	return p
}
