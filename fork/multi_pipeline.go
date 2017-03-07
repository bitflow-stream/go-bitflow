package fork

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

type MultiPipeline struct {
	parallelClose bool

	pipelines        []*runningSubPipeline
	runningPipelines int
	stopped          bool
	stoppedCond      *sync.Cond
	subPipelineWg    sync.WaitGroup
	merger           ForkMerger
}

func (m *MultiPipeline) Init(outgoing bitflow.MetricSink, closeHook func(), wg *sync.WaitGroup) {
	m.stoppedCond = sync.NewCond(new(sync.Mutex))
	m.merger.outgoing = outgoing
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.waitForPipelines()
		closeHook()
	}()
}

func (m *MultiPipeline) StartPipeline(pipeline *bitflow.SamplePipeline, finishedHook func(isPassive bool, err error)) {
	if pipeline.Sink == nil {
		pipeline.Sink = &m.merger
	}
	if pipeline.Source == nil {
		// Use an empty source to make stopPipeline() work
		pipeline.Source = new(bitflow.EmptyMetricSource)
	}
	m.runningPipelines++

	running := runningSubPipeline{
		pipeline: pipeline,
	}
	m.pipelines = append(m.pipelines, &running)
	tasks, channels := running.init(&m.subPipelineWg)

	m.subPipelineWg.Add(1)
	go func() {
		defer m.subPipelineWg.Done()

		// Wait for all tasks to finish and collect their errors
		idx := golib.WaitForAny(channels)
		errors := tasks.CollectMultiError(channels)

		// A passive pipeline can occur when all processors, source and sink return nil from Start().
		// This means that none of the elements of the sub pipeline spawned any extra goroutines.
		// They only react on Sample() and wait for the final Close() call.
		isPassive := idx == -1
		finishedHook(isPassive, errors.NilOrError())

		m.stoppedCond.L.Lock()
		defer m.stoppedCond.L.Unlock()
		m.runningPipelines--
		m.stoppedCond.Broadcast()
	}()
	return
}

func (m *MultiPipeline) StopPipelines() {
	m.stopPipelines()

	m.stoppedCond.L.Lock()
	defer m.stoppedCond.L.Unlock()
	m.stopped = true
	m.stoppedCond.Broadcast()
}

func (m *MultiPipeline) stopPipelines() {
	var wg sync.WaitGroup
	for i, pipeline := range m.pipelines {
		m.pipelines[i] = nil // Enable GC
		if pipeline != nil {
			wg.Add(1)
			go func(pipeline *runningSubPipeline) {
				defer wg.Done()
				pipeline.stop()
			}(pipeline)
			if !m.parallelClose {
				wg.Wait()
			}
		}
	}
	wg.Wait()
}

func (m *MultiPipeline) waitForPipelines() {
	m.stoppedCond.L.Lock()
	defer m.stoppedCond.L.Unlock()
	for !m.stopped || m.runningPipelines > 0 {
		m.stoppedCond.Wait()
	}
	m.subPipelineWg.Wait()
}

func (m *MultiPipeline) LogFinishedPipeline(isPassive bool, err error, prefix string) {
	if isPassive {
		prefix += " is passive"
	} else {
		prefix += " finished"
	}
	if err == nil {
		log.Debugln(prefix)
	} else {
		log.Errorf("%v (Error: %v)", prefix, err)
	}
}

type runningSubPipeline struct {
	pipeline *bitflow.SamplePipeline
	group    golib.TaskGroup
}

func (r *runningSubPipeline) init(wg *sync.WaitGroup) (golib.TaskGroup, []golib.StopChan) {
	r.pipeline.Construct(&r.group)
	return r.group, r.group.StartTasks(wg)
}

func (r *runningSubPipeline) stop() {
	r.group.Stop()
}

type ForkMerger struct {
	bitflow.AbstractMetricSink
	mutex    sync.Mutex
	outgoing bitflow.MetricSink
}

func (sink *ForkMerger) String() string {
	return "Fork merger for " + sink.outgoing.String()
}

func (sink *ForkMerger) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	return
}

func (sink *ForkMerger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	sink.mutex.Lock()
	defer sink.mutex.Unlock()
	return sink.outgoing.Sample(sample, header)
}

func (sink *ForkMerger) Close() {
	// The actual outgoing sink must be closed after waitForPipelines() returns
}

// Can be used by implementations of PipelineBuilder to access the next step of the entire fork.
func (sink *ForkMerger) GetOriginalSink() bitflow.MetricSink {
	return sink.outgoing
}
