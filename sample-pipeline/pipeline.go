package pipeline

import (
	"sync"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

// Combination of sample.MetricSink and sample.MetricSource
type SampleProcessor interface {
	golib.Task
	Header(header sample.Header) error
	Sample(sample sample.Sample, header sample.Header) error
	SetSink(sink sample.MetricSink)
	Close()
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

func (p *SamplePipeline) Add(processor SampleProcessor) *SamplePipeline {
	p.Processors = append(p.Processors, processor)
	return p
}

// ==================== Abstract Processor ====================
// This type defines standard implementations for all methods in
// SampleProcessor. Parts can be shadowed by embedding the type.
type AbstractProcessor struct {
	sample.AbstractMetricSource
	sample.AbstractMetricSink
}

func (p *AbstractProcessor) Header(header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		return p.OutgoingSink.Header(header)
	}
}

func (p *AbstractProcessor) Sample(sample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	return p.OutgoingSink.Sample(sample, header)
}

func (p *AbstractProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	return nil
}

func (p *AbstractProcessor) Close() {
	// Propagate the Close() invocation
	p.CloseSink(nil)
}

func (p *AbstractProcessor) String() string {
	return "AbstractProcessor"
}
