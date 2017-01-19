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

	pipelines        map[interface{}]bitflow.MetricSink
	runningPipelines int
	stopped          bool
	stoppedCond      sync.Cond
	subpipelineWg    sync.WaitGroup
	merger           ForkMerger
}

func NewMetricFork(distributor ForkDistributor, builder PipelineBuilder) *MetricFork {
	fork := &MetricFork{
		Builder:       builder,
		Distributor:   distributor,
		pipelines:     make(map[interface{}]bitflow.MetricSink),
		ParallelClose: true,
	}
	fork.stoppedCond.L = new(sync.Mutex)
	fork.merger.fork = fork
	return fork
}

func (f *MetricFork) Start(wg *sync.WaitGroup) golib.StopChan {
	result := f.AbstractProcessor.Start(wg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.waitForSubpipelines()
		f.CloseSink()
	}()
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
	f.stoppedCond.L.Lock()
	defer f.stoppedCond.L.Unlock()
	f.stopped = true

	var wg sync.WaitGroup
	for i, pipeline := range f.pipelines {
		f.pipelines[i] = nil // Enable GC
		if pipeline != nil {
			wg.Add(1)
			go func(pipeline bitflow.MetricSink) {
				defer wg.Done()
				pipeline.Close()
			}(pipeline)
			if !f.ParallelClose {
				wg.Wait()
			}
		}
	}
	wg.Wait()
	f.stoppedCond.Broadcast()
}

func (f *MetricFork) String() string {
	return fmt.Sprintf("Fork. Distributor: %v, Builder: %v", f.Distributor, f.Builder)
}

func (f *MetricFork) newPipeline(key interface{}) bitflow.MetricSink {
	log.Debugf("[%v]: Starting forked subpipeline %v", f, key)
	pipeline := f.Builder.BuildPipeline(key, &f.merger)
	group := golib.NewTaskGroup()
	pipeline.Source = nil // The source is already started
	pipeline.Construct(group)
	f.runningPipelines++
	waitingTasks, channels := group.StartTasks(&f.subpipelineWg)
	f.subpipelineWg.Add(1)

	go func() {
		defer f.subpipelineWg.Done()

		idx, err := golib.WaitForAny(channels)
		if err != nil {
			log.Errorln(err)
		}
		group.ReverseStop(golib.DefaultPrintTaskStopWait) // The Stop() calls are actually ignored because group contains only MetricSinks
		_ = golib.PrintErrors(channels, waitingTasks, golib.DefaultPrintTaskStopWait)

		if idx == -1 {
			// Inactive subpipeline can occur when all processors and the sink return nil from Start().
			// This means that none of the elements of the subpipeline spawned any extra goroutines.
			// They only react on Sample() and wait for the final Close() call.
			log.Debugf("[%v]: Subpipeline inactive: %v", f, key)
		} else {
			log.Debugf("[%v]: Finished forked subpipeline %v", f, key)
		}

		f.stoppedCond.L.Lock()
		defer f.stoppedCond.L.Unlock()
		f.runningPipelines--
		f.stoppedCond.Broadcast()
	}()

	var first bitflow.MetricSink
	if len(pipeline.Processors) == 0 {
		first = pipeline.Sink
	} else {
		first = pipeline.Processors[0]
	}
	f.pipelines[key] = first
	return first
}

func (f *MetricFork) waitForSubpipelines() {
	f.stoppedCond.L.Lock()
	for !f.stopped || f.runningPipelines > 0 {
		f.stoppedCond.Wait()
	}
	f.stoppedCond.L.Unlock()
	f.subpipelineWg.Wait()
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
