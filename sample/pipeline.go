package sample

import "github.com/antongulenko/golib"

// Combination of MetricSink and MetricSource
type SampleProcessor interface {
	golib.Task
	Header(header Header) error
	Sample(sample Sample, header Header) error
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
	source := p.Source
	tasks.Add(source)
	for _, processor := range p.Processors {
		if source != nil {
			source.SetSink(processor)
		}
		source = processor
		tasks.Add(source)
	}
	if source != nil {
		source.SetSink(p.Sink)
	}

	if agg, ok := p.Sink.(AggregateSink); ok {
		for _, sink := range agg {
			tasks.Add(sink)
		}
	} else {
		tasks.Add(p.Sink)
	}
}

func (p *SamplePipeline) Add(processor SampleProcessor) *SamplePipeline {
	p.Processors = append(p.Processors, processor)
	return p
}
