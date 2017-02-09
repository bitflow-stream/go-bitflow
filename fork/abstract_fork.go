package fork

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/golib"
)

type PipelineBuilder interface {
	BuildPipeline(key interface{}, output *ForkMerger) *bitflow.SamplePipeline
	String() string
}

type AbstractMetricFork struct {
	MultiPipeline
	bitflow.AbstractProcessor
	pipelines map[interface{}]bitflow.MetricSink
	lock      sync.Mutex
}

func (f *AbstractMetricFork) Start(wg *sync.WaitGroup) golib.StopChan {
	result := f.AbstractProcessor.Start(wg)
	f.MultiPipeline.Init(f.OutgoingSink, f.CloseSink, wg)
	f.pipelines = make(map[interface{}]bitflow.MetricSink)
	return result
}

func (f *AbstractMetricFork) getPipeline(builder PipelineBuilder, key interface{}, description fmt.Stringer) bitflow.MetricSink {
	f.lock.Lock()
	defer f.lock.Unlock()
	pipeline, ok := f.pipelines[key]
	if !ok {
		pipeline = f.newPipeline(builder, key, description)
		f.pipelines[key] = pipeline
	}
	return pipeline
}

func (f *AbstractMetricFork) newPipeline(builder PipelineBuilder, key interface{}, description fmt.Stringer) bitflow.MetricSink {
	log.Debugf("[%v]: Starting forked subpipeline %v", description, key)
	pipeline := builder.BuildPipeline(key, &f.merger)
	if pipeline.Source != nil {
		// Forked pipelines should not have an explicit source, as they receive
		// samples from the steps preceeding them
		log.Warnf("[%v]: The Source field of the %v subpipeline was set and will be ignored: %v", description, key, pipeline.Source)
		pipeline.Source = nil
	}
	f.StartPipeline(pipeline, func(isPassive bool, err error) {
		f.LogFinishedPipeline(isPassive, err, fmt.Sprintf("[%v]: Subpipeline %v", description, key))
	})

	var first bitflow.MetricSink
	if len(pipeline.Processors) == 0 {
		first = pipeline.Sink
	} else {
		first = pipeline.Processors[0]
	}
	return first
}

func (f *AbstractMetricFork) containedStringers(builder PipelineBuilder) []fmt.Stringer {
	if container, ok := builder.(pipeline.StringerContainer); ok {
		return container.ContainedStringers()
	} else {
		return []fmt.Stringer{builder}
	}
}

func (f *AbstractMetricFork) Close() {
	f.StopPipelines()
}
