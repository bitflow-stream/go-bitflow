package fork

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/golib"
)

type ForkDistributor interface {
	Distribute(sample *bitflow.Sample, header *bitflow.Header) []interface{}
	String() string
}

type PipelineBuilder interface {
	BuildPipeline(key interface{}, output *ForkMerger) *bitflow.SamplePipeline
	String() string
}

type MetricFork struct {
	MultiPipeline
	bitflow.AbstractProcessor

	Distributor ForkDistributor
	Builder     PipelineBuilder

	pipelines map[interface{}]bitflow.MetricSink
}

func (f *MetricFork) Start(wg *sync.WaitGroup) golib.StopChan {
	result := f.AbstractProcessor.Start(wg)
	f.MultiPipeline.Init(f.OutgoingSink, f.CloseSink, wg)
	return result
}

func (f *MetricFork) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := f.Check(sample, header); err != nil {
		return err
	}
	keys := f.Distributor.Distribute(sample, header)
	var errors golib.MultiError
	for _, key := range keys {
		pipeline, ok := f.pipelines[key]
		if !ok {
			pipeline = f.newPipeline(key)
		}
		if pipeline != nil {
			errors.Add(pipeline.Sample(sample, header))
		}
	}
	return errors.NilOrError()
}

func (f *MetricFork) newPipeline(key interface{}) bitflow.MetricSink {
	if f.pipelines == nil {
		f.pipelines = make(map[interface{}]bitflow.MetricSink)
	}

	log.Debugf("[%v]: Starting forked subpipeline %v", f, key)
	pipeline := f.Builder.BuildPipeline(key, &f.merger)
	if pipeline.Source != nil {
		// Forked pipelines should not have an explicit source, as they receive
		// samples from the steps preceeding them
		log.Warnf("The Source field of the %v subpipeline was set and will be ignored: %v", key, pipeline.Source)
		pipeline.Source = nil
	}
	f.StartPipeline(pipeline, func(isPassive bool, err error) {
		f.LogFinishedPipeline(isPassive, err, fmt.Sprintf("[%v]: Subpipeline %v", f, key))
	})

	var first bitflow.MetricSink
	if len(pipeline.Processors) == 0 {
		first = pipeline.Sink
	} else {
		first = pipeline.Processors[0]
	}
	f.pipelines[key] = first
	return first
}

func (f *MetricFork) Close() {
	f.StopPipelines()
}

func (f *MetricFork) String() string {
	return "Fork: " + f.Distributor.String()
}

func (f *MetricFork) ContainedStringers() []fmt.Stringer {
	if builder, ok := f.Builder.(pipeline.StringerContainer); ok {
		return builder.ContainedStringers()
	} else {
		return []fmt.Stringer{f.Builder}
	}
}

type SimplePipelineBuilder struct {
	Build           func() []bitflow.SampleProcessor
	examplePipeline []fmt.Stringer
}

func (b *SimplePipelineBuilder) ContainedStringers() []fmt.Stringer {
	if b.examplePipeline == nil {
		if b.Build == nil {
			b.examplePipeline = make([]fmt.Stringer, 0)
		} else {
			pipeline := b.Build()
			b.examplePipeline = make([]fmt.Stringer, len(pipeline))
			for i, step := range pipeline {
				b.examplePipeline[i] = step
			}
		}
	}
	return b.examplePipeline
}

func (b *SimplePipelineBuilder) String() string {
	return fmt.Sprintf("Simple Pipeline Builder len %v", len(b.ContainedStringers()))
}

func (b *SimplePipelineBuilder) BuildPipeline(key interface{}, output *ForkMerger) *bitflow.SamplePipeline {
	var res bitflow.SamplePipeline
	res.Sink = output
	if b.Build != nil {
		for _, processor := range b.Build() {
			res.Add(processor)
		}
	}
	return &res
}
