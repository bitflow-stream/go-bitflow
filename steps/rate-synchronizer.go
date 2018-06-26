package steps

import (
	"sync"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
)

type PipelineRateSynchronizer struct {
	steps     []*synchronizationStep
	startOnce sync.Once

	ChannelSize      int
	ChannelCloseHook func(lastSample *bitflow.Sample, lastHeader *bitflow.Header)
}

func RegisterPipelineRateSynchronizer(b *query.PipelineBuilder) {
	synchronization_keys := make(map[string]*PipelineRateSynchronizer)

	create := func(p *pipeline.SamplePipeline, params map[string]string) error {
		key := params["key"]
		synchronizer, ok := synchronization_keys[key]
		if !ok {
			synchronizer = &PipelineRateSynchronizer{
				ChannelSize: 5, // TODO parameterize
			}
			synchronization_keys[key] = synchronizer
		}
		p.Add(synchronizer.NewSynchronizationStep())
		return nil
	}

	b.RegisterAnalysisParamsErr("synchronize", create, "Synchronize the number of samples going through each synchronize() step with the same key parameter", []string{"key"})
}

func (s *PipelineRateSynchronizer) NewSynchronizationStep() bitflow.SampleProcessor {
	chanSize := s.ChannelSize
	if chanSize < 1 {
		chanSize = 1
	}
	step := &synchronizationStep{
		synchronizer: s,
		queue:        make(chan sampleAndHeader, chanSize),
		running:      true,
	}
	s.steps = append(s.steps, step)
	return step
}

func (s *PipelineRateSynchronizer) start(wg *sync.WaitGroup) {
	s.startOnce.Do(func() {
		wg.Add(1)
		go s.process(wg)
	})
}

func (s *PipelineRateSynchronizer) process(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		runningSteps := 0
		for _, step := range s.steps {
			if step.running {
				if sample := <-step.queue; sample.valid {
					step.outputSample(sample)
					runningSteps++
				} else {
					step.running = false
					step.CloseSink()
				}
			}
		}
		if runningSteps == 0 {
			break
		}
	}
}

type sampleAndHeader struct {
	sample *bitflow.Sample
	header *bitflow.Header
	valid  bool
}

type synchronizationStep struct {
	bitflow.NoopProcessor
	synchronizer  *PipelineRateSynchronizer
	queue         chan sampleAndHeader
	running       bool
	closeSinkOnce sync.Once
	err           error
	lastSample    *bitflow.Sample
	lastHeader    *bitflow.Header
}

func (s *synchronizationStep) String() string {
	return "Synchronize processing rate"
}

func (s *synchronizationStep) Start(wg *sync.WaitGroup) golib.StopChan {
	s.synchronizer.start(wg)
	return s.NoopProcessor.Start(wg)
}

func (s *synchronizationStep) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if e := s.err; e != nil {
		s.err = nil
		return e
	}
	s.lastSample = sample
	s.lastHeader = header
	s.queue <- sampleAndHeader{
		sample: sample,
		header: header,
		valid:  true,
	}
	return nil
}

func (s *synchronizationStep) Close() {
	close(s.queue)
}

func (s *synchronizationStep) CloseSink() {
	s.closeSinkOnce.Do(func() {
		if c := s.synchronizer.ChannelCloseHook; c != nil {
			c(s.lastSample, s.lastHeader)
		}
		s.NoopProcessor.CloseSink()
	})
}

func (s *synchronizationStep) outputSample(sample sampleAndHeader) {
	err := s.NoopProcessor.Sample(sample.sample, sample.header)
	if err != nil && s.err == nil {
		s.err = err
	}
}
