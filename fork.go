package pipeline

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
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
	bitflow.AbstractProcessor

	Distributor   ForkDistributor
	Builder       PipelineBuilder
	ParallelClose bool

	pipelines map[interface{}]bitflow.MetricSink
	multi     *MultiPipeline
	merger    ForkMerger
}

func NewMetricFork(distributor ForkDistributor, builder PipelineBuilder) *MetricFork {
	fork := &MetricFork{
		Builder:       builder,
		Distributor:   distributor,
		pipelines:     make(map[interface{}]bitflow.MetricSink),
		ParallelClose: true,
		multi:         NewMultiPipeline(),
	}
	fork.merger.fork = fork
	return fork
}

func (f *MetricFork) Start(wg *sync.WaitGroup) golib.StopChan {
	result := f.AbstractProcessor.Start(wg)
	f.multi.RunAfterClose(f.CloseSink, wg)
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

func (f *MetricFork) Close() {
	f.multi.Close()
}

func (f *MetricFork) String() string {
	return "Fork: " + f.Distributor.String()
}

func (f *MetricFork) ContainedStringers() []fmt.Stringer {
	if builder, ok := f.Builder.(StringerContainer); ok {
		return builder.ContainedStringers()
	} else {
		return []fmt.Stringer{f.Builder}
	}
}

func (f *MetricFork) newPipeline(key interface{}) bitflow.MetricSink {
	log.Debugf("[%v]: Starting forked subpipeline %v", f, key)
	pipeline := f.Builder.BuildPipeline(key, &f.merger)
	if pipeline.Source != nil {
		// Forked pipelines should not have an explicit source, as they receive
		// samples from the steps preceeding them
		log.Warnf("The Source field of the %v subpipeline was set (%v) and is ignored", key, pipeline.Source)
		pipeline.Source = nil
	}
	first := f.multi.StartPipeline(pipeline, func(isPassive bool) {
		f.pipelines[key] = nil // Allow GC
		if isPassive {
			log.Debugf("[%v]: Passive subpipeline: %v", f, key)
		} else {
			log.Debugf("[%v]: Finished forked subpipeline %v", f, key)
		}
	})
	f.pipelines[key] = first
	return first
}

type ForkMerger struct {
	bitflow.AbstractMetricSink
	mutex sync.Mutex
	fork  *MetricFork
}

func (sink *ForkMerger) String() string {
	return "fork merger for " + sink.fork.String()
}

func (sink *ForkMerger) Start(wg *sync.WaitGroup) golib.StopChan {
	return nil
}

func (sink *ForkMerger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	sink.mutex.Lock()
	defer sink.mutex.Unlock()
	return sink.fork.OutgoingSink.Sample(sample, header)
}

func (sink *ForkMerger) Close() {
	// The actual outgoing sink is closed after waitForSubpipelines() returns
}

// Can be used by implementations of PipelineBuilder to access the next step of the entire fork.
func (sink *ForkMerger) GetOriginalSink() bitflow.MetricSink {
	return sink.fork.OutgoingSink
}

type SimplePipelineBuilder struct {
	Build           func() []bitflow.SampleProcessor
	examplePipeline []bitflow.SampleProcessor
}

func (b *SimplePipelineBuilder) String() string {
	if b.examplePipeline == nil {
		if b.Build == nil {
			b.examplePipeline = make([]bitflow.SampleProcessor, 0)
		} else {
			b.examplePipeline = b.Build()
		}
	}
	if len(b.examplePipeline) == 0 {
		return "(empty subpipeline)"
	} else {
		return fmt.Sprintf("subpipeline %v", b.examplePipeline)
	}
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
