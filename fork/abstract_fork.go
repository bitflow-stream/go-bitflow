package fork

import (
	"fmt"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

type PipelineBuilder interface {
	BuildPipeline(key interface{}, output *ForkMerger) *bitflow.SamplePipeline
	String() string
}

type AbstractMetricFork struct {
	MultiPipeline
	bitflow.NoopProcessor
	pipelines map[interface{}]bitflow.SampleProcessor
	lock      sync.Mutex

	newPipelineHandler func(bitflow.SampleProcessor) bitflow.SampleProcessor // Optional hook
	ForkPath           []interface{}
}

func (f *AbstractMetricFork) Start(wg *sync.WaitGroup) golib.StopChan {
	result := f.NoopProcessor.Start(wg)
	f.MultiPipeline.Init(f.GetSink(), f.CloseSink, wg)
	f.pipelines = make(map[interface{}]bitflow.SampleProcessor)
	return result
}

func (f *AbstractMetricFork) Close() {
	f.StopPipelines()
}

func (f *AbstractMetricFork) getPipelines(builder PipelineBuilder, keys []interface{}, description fmt.Stringer) bitflow.SampleProcessor {
	sinks := make([]bitflow.SampleProcessor, len(keys))
	for i, key := range keys {
		sinks[i] = f.getPipeline(builder, key, description)
	}
	return &multiPipelineSink{sinks: sinks}
}

func (f *AbstractMetricFork) getPipeline(builder PipelineBuilder, key interface{}, description fmt.Stringer) bitflow.SampleProcessor {
	f.lock.Lock()
	defer f.lock.Unlock()
	pipe, ok := f.pipelines[key]
	if !ok {
		pipe = f.newPipeline(builder, key, description)
		if hook := f.newPipelineHandler; hook != nil {
			pipe = hook(pipe)
		}
		f.pipelines[key] = pipe
	}
	return pipe
}

func (f *AbstractMetricFork) newPipeline(builder PipelineBuilder, key interface{}, description fmt.Stringer) bitflow.SampleProcessor {
	pipe := builder.BuildPipeline(key, &f.merger)
	path := f.setForkPaths(pipe, key)
	log.Debugf("[%v]: Starting forked subpipeline %v", description, path)
	if pipe.Source != nil {
		// Forked pipelines should not have an explicit source, as they receive
		// samples from the steps preceding them
		log.Warnf("[%v]: The Source field of the %v subpipeline was set and will be ignored: %v", description, path, pipe.Source)
		pipe.Source = nil
	}
	// Special handling of ForkRemapper: automatically connect mapped pipelines
	pipe.Add(f.getRemappedSink(pipe, path))
	f.StartPipeline(pipe, func(isPassive bool, err error) {
		f.LogFinishedPipeline(isPassive, err, fmt.Sprintf("[%v]: Subpipeline %v", description, path))
	})
	return pipe.Processors[0]
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

func (f *AbstractMetricFork) getRemappedSink(pipeline *bitflow.SamplePipeline, forkPath []interface{}) bitflow.SampleProcessor {
	if len(pipeline.Processors) > 0 {
		last := pipeline.Processors[len(pipeline.Processors)-1]
		if _, isFork := last.(abstractForkContainer); isFork {
			// If the last step is a fork, it will handle remapping on its own
			return nil
		}
	}
	return f.getRemappedSinkRecursive(f.GetSink(), forkPath)
}

func (f *AbstractMetricFork) getRemappedSinkRecursive(outgoing bitflow.SampleProcessor, forkPath []interface{}) bitflow.SampleProcessor {
	switch outgoing := outgoing.(type) {
	case *ForkRemapper:
		// Ask follow-up ForkRemapper for the pipeline we should connect to
		return outgoing.GetMappedSink(forkPath)
	case *ForkMerger:
		// If there are multiple layers of forks, we have to resolve the ForkMergers until we get the actual outgoing sink
		return f.getRemappedSinkRecursive(outgoing.GetOriginalSink(), forkPath)
	default:
		// No follow-up ForkRemapper could be found
		return nil
	}
}

type abstractForkContainer interface {
	getAbstractFork() *AbstractMetricFork
}

func (f *AbstractMetricFork) getAbstractFork() *AbstractMetricFork {
	return f
}

type multiPipelineSink struct {
	bitflow.DroppingSampleProcessor
	sinks []bitflow.SampleProcessor
}

func (s *multiPipelineSink) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	var errors golib.MultiError
	for _, sink := range s.sinks {
		if sink != nil {
			// The DeepClone() is necessary since the forks might change the sample
			// values independently. In some cases it might not be necessary, but that
			// would be a rather complex optimization.
			errors.Add(sink.Sample(sample.DeepClone(), header))
		}
	}
	return errors.NilOrError()
}

func (s *multiPipelineSink) String() string {
	return fmt.Sprintf("parallel multi sink len %v", len(s.sinks))
}
