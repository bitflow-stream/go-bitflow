package steps

import (
	"errors"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

type PipelineRateSynchronizer struct {
	steps     []*synchronizationStep
	startOnce sync.Once

	ChannelSize      int
	ChannelCloseHook func(lastSample *bitflow.Sample, lastHeader *bitflow.Header)
}

func RegisterPipelineRateSynchronizer(b reg.ProcessorRegistry) {
	synchronization_keys := make(map[string]*PipelineRateSynchronizer)

	create := func(p *bitflow.SamplePipeline, params map[string]string) error {
		var err error
		key := params["key"]
		chanSize := reg.IntParam(params, "buf", 5, true, &err)
		if err != nil {
			return err
		}

		synchronizer, ok := synchronization_keys[key]
		if !ok {
			synchronizer = &PipelineRateSynchronizer{
				ChannelSize: chanSize,
			}
			synchronization_keys[key] = synchronizer
		} else if synchronizer.ChannelSize != chanSize {
			return reg.ParameterError("buf", errors.New("synchronize() steps with the same 'key' parameter must all have the same 'buf' parameter"))
		}
		p.Add(synchronizer.NewSynchronizationStep())
		return nil
	}

	b.RegisterAnalysisParamsErr("synchronize", create, "Synchronize the number of samples going through each synchronize() step with the same key parameter", reg.RequiredParams("key"), reg.OptionalParams("buf"))
}

func (s *PipelineRateSynchronizer) NewSynchronizationStep() bitflow.SampleProcessor {
	chanSize := s.ChannelSize
	if chanSize < 1 {
		chanSize = 1
	}
	step := &synchronizationStep{
		synchronizer: s,
		queue:        make(chan bitflow.SampleAndHeader, chanSize),
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
				if sample := <-step.queue; sample.Sample != nil {
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

type synchronizationStep struct {
	bitflow.NoopProcessor
	synchronizer  *PipelineRateSynchronizer
	queue         chan bitflow.SampleAndHeader
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
	s.queue <- bitflow.SampleAndHeader{
		Sample: sample,
		Header: header,
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

func (s *synchronizationStep) outputSample(sample bitflow.SampleAndHeader) {
	err := s.NoopProcessor.Sample(sample.Sample, sample.Header)
	if err != nil && s.err == nil {
		s.err = err
	}
}
