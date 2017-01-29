package pipeline

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

type MultiPipeline struct {
	ParallelClose bool

	pipelines        []bitflow.MetricSink
	runningPipelines int
	stopped          bool
	stoppedCond      *sync.Cond
	subpipelineWg    sync.WaitGroup
}

func NewMultiPipeline() *MultiPipeline {
	return &MultiPipeline{
		ParallelClose: true,
		stoppedCond:   sync.NewCond(new(sync.Mutex)),
	}
}

func (f *MultiPipeline) RunAfterClose(closeHook func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.waitForPipelines()
		closeHook()
	}()
}

func (m *MultiPipeline) StartPipeline(pipeline *bitflow.SamplePipeline, finishedHook func(isPassive bool)) (first bitflow.MetricSink) {
	if len(pipeline.Processors) == 0 {
		first = pipeline.Sink
	} else {
		first = pipeline.Processors[0]
	}
	m.pipelines = append(m.pipelines, first)
	m.runningPipelines++

	group := golib.NewTaskGroup()
	pipeline.Construct(group)
	waitingTasks, channels := group.StartTasks(&m.subpipelineWg)

	m.subpipelineWg.Add(1)
	go func() {
		defer m.subpipelineWg.Done()

		idx, err := golib.WaitForAny(channels)
		if err != nil {
			log.Errorln(err)
		}
		group.ReverseStop(golib.DefaultPrintTaskStopWait) // The Stop() calls are actually ignored because group contains only MetricSinks
		_ = golib.PrintErrors(channels, waitingTasks, golib.DefaultPrintTaskStopWait)

		if finishedHook != nil {
			// A passive pipeline can occur when all processors, source and sink return nil from Start().
			// This means that none of the elements of the subpipeline spawned any extra goroutines.
			// They only react on Sample() and wait for the final Close() call.
			isPassive := idx == -1
			finishedHook(isPassive)
		}
		m.stoppedCond.L.Lock()
		defer m.stoppedCond.L.Unlock()
		m.runningPipelines--
		m.stoppedCond.Broadcast()
	}()
	return
}

func (m *MultiPipeline) Close() {
	m.stoppedCond.L.Lock()
	defer m.stoppedCond.L.Unlock()
	m.stopped = true

	var wg sync.WaitGroup
	for i, pipeline := range m.pipelines {
		m.pipelines[i] = nil // Enable GC
		if pipeline != nil {
			wg.Add(1)
			go func(pipeline bitflow.MetricSink) {
				defer wg.Done()
				pipeline.Close()
			}(pipeline)
			if !m.ParallelClose {
				wg.Wait()
			}
		}
	}
	wg.Wait()
	m.stoppedCond.Broadcast()
}

func (m *MultiPipeline) waitForPipelines() {
	m.stoppedCond.L.Lock()
	for !m.stopped || m.runningPipelines > 0 {
		m.stoppedCond.Wait()
	}
	m.stoppedCond.L.Unlock()
	m.subpipelineWg.Wait()
}
