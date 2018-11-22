package fork

import (
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	log "github.com/sirupsen/logrus"
)

type MultiPipeline struct {
	SequentialClose bool

	pipelines        []*runningSubPipeline
	runningPipelines int
	stopped          bool
	stoppedCond      *sync.Cond
	subPipelineWg    sync.WaitGroup
	merger           Merger
}

func (m *MultiPipeline) Init(outgoing bitflow.SampleProcessor, closeHook func(), wg *sync.WaitGroup) {
	m.stoppedCond = sync.NewCond(new(sync.Mutex))
	m.merger.outgoing = outgoing
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.waitForStoppedPipelines()
		m.subPipelineWg.Wait()
		closeHook()
	}()
}

func (m *MultiPipeline) StartPipeline(pipeline *bitflow.SamplePipeline, finishedHook func(isPassive bool, err error)) {
	pipeline.Add(&m.merger)
	if pipeline.Source == nil {
		// Use an empty source to make stopPipeline() work
		pipeline.Source = new(bitflow.EmptySampleSource)
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
			if m.SequentialClose {
				wg.Wait()
			}
		}
	}
	wg.Wait()
}

func (m *MultiPipeline) waitForStoppedPipelines() {
	m.stoppedCond.L.Lock()
	defer m.stoppedCond.L.Unlock()
	for !m.stopped || m.runningPipelines > 0 {
		m.stoppedCond.Wait()
	}
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

type Merger struct {
	bitflow.AbstractSampleProcessor
	mutex    sync.Mutex
	outgoing bitflow.SampleProcessor
}

func (sink *Merger) String() string {
	return "Fork merger for " + sink.outgoing.String()
}

func (sink *Merger) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	return
}

func (sink *Merger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	sink.mutex.Lock()
	defer sink.mutex.Unlock()
	return sink.outgoing.Sample(sample, header)
}

func (sink *Merger) Close() {
	// The actual outgoing sink must be closed in the closeHook function passed to Init()
}
