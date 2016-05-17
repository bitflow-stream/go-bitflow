package pipeline

import (
	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

// Combination of sample.MetricSink and sample.MetricSource
type SampleProcessor interface {
	golib.Task
	Header(header sample.Header) error
	Sample(sample sample.Sample, header sample.Header) error
	SetSink(sink sample.MetricSink)
}

type SamplePipeline struct {
	Source     sample.MetricSource
	Processors []SampleProcessor
	Sink       sample.MetricSink
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

	if agg, ok := p.Sink.(sample.AggregateSink); ok {
		for _, sink := range agg {
			tasks.Add(sink)
		}
	} else {
		tasks.Add(p.Sink)
	}
}

func (p *SamplePipeline) Add(processor SampleProcessor) {
	p.Processors = append(p.Processors, processor)
}
