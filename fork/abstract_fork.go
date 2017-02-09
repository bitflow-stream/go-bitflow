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

	newPipelineHandler func(bitflow.MetricSink) bitflow.MetricSink // Optional hook
	ForkPath           []interface{}
}

func (f *AbstractMetricFork) Start(wg *sync.WaitGroup) golib.StopChan {
	result := f.AbstractProcessor.Start(wg)
	f.MultiPipeline.Init(f.OutgoingSink, f.CloseSink, wg)
	f.pipelines = make(map[interface{}]bitflow.MetricSink)
	return result
}

func (f *AbstractMetricFork) Close() {
	f.StopPipelines()
}

func (f *AbstractMetricFork) getPipelines(builder PipelineBuilder, keys []interface{}, description fmt.Stringer) bitflow.MetricSink {
	sinks := make([]bitflow.MetricSink, len(keys))
	for i, key := range keys {
		sinks[i] = f.getPipeline(builder, key, description)
	}
	return &multiPipelineSink{sinks: sinks}
}

func (f *AbstractMetricFork) getPipeline(builder PipelineBuilder, key interface{}, description fmt.Stringer) bitflow.MetricSink {
	f.lock.Lock()
	defer f.lock.Unlock()
	pipeline, ok := f.pipelines[key]
	if !ok {
		pipeline = f.newPipeline(builder, key, description)
		if hook := f.newPipelineHandler; hook != nil {
			pipeline = hook(pipeline)
		}
		f.pipelines[key] = pipeline
	}
	return pipeline
}

func (f *AbstractMetricFork) newPipeline(builder PipelineBuilder, key interface{}, description fmt.Stringer) bitflow.MetricSink {
	pipeline := builder.BuildPipeline(key, &f.merger)
	path := f.setForkPaths(pipeline, key)
	log.Debugf("[%v]: Starting forked subpipeline %v", description, path)
	if pipeline.Source != nil {
		// Forked pipelines should not have an explicit source, as they receive
		// samples from the steps preceeding them
		log.Warnf("[%v]: The Source field of the %v subpipeline was set and will be ignored: %v", description, path, pipeline.Source)
		pipeline.Source = nil
	}
	if pipeline.Sink == nil {
		// Special handling of ForkRemapper: automatically connect mapped pipelines
		if remapper, ok := f.OutgoingSink.(*ForkRemapper); ok {
			pipeline.Sink = remapper.GetMappedSink(path)
		}
	}
	f.StartPipeline(pipeline, func(isPassive bool, err error) {
		f.LogFinishedPipeline(isPassive, err, fmt.Sprintf("[%v]: Subpipeline %v", description, path))
	})

	if len(pipeline.Processors) == 0 {
		return pipeline.Sink
	} else {
		return pipeline.Processors[0]
	}
}

func (f *AbstractMetricFork) containedStringers(builder PipelineBuilder) []fmt.Stringer {
	if container, ok := builder.(pipeline.StringerContainer); ok {
		return container.ContainedStringers()
	} else {
		return []fmt.Stringer{builder}
	}
}

func (f *AbstractMetricFork) setForkPaths(pipeline *bitflow.SamplePipeline, key interface{}) []interface{} {
	path := make([]interface{}, len(f.ForkPath)+1)
	copy(path, f.ForkPath)
	path[len(path)-1] = key
	for _, proc := range pipeline.Processors {
		if forkContainer, ok := proc.(abstractForkContainer); ok {
			forkContainer.getAbstractFork().ForkPath = path
		}
	}
	return path
}

type abstractForkContainer interface {
	getAbstractFork() *AbstractMetricFork
}

func (f *AbstractMetricFork) getAbstractFork() *AbstractMetricFork {
	return f
}

type multiPipelineSink struct {
	bitflow.EmptyMetricSink
	sinks []bitflow.MetricSink
}

func (s *multiPipelineSink) Sample(sample *bitflow.Sample, header *bitflow.Header) error {

	// TODO implement optional parallelism

	var errors golib.MultiError
	for _, sink := range s.sinks {
		if sink != nil {
			errors.Add(sink.Sample(sample, header))
		}
	}
	return errors.NilOrError()
}

func (s *multiPipelineSink) String() string {
	return fmt.Sprintf("parallel multi sink len %v", len(s.sinks))
}
