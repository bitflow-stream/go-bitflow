package pipeline

import (
	"fmt"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

type MultiMetricSource struct {
	MultiPipeline
	bitflow.AbstractMetricSource

	pipelines        []*SamplePipeline
	stoppedPipelines int
}

func (in *MultiMetricSource) Add(subpipeline *SamplePipeline) {
	in.pipelines = append(in.pipelines, subpipeline)
}

func (in *MultiMetricSource) Start(wg *sync.WaitGroup) golib.StopChan {
	stopChan := make(chan error, 2)
	signalClose := func() {
		in.CloseSink(wg)
		stopChan <- nil
	}

	in.MultiPipeline.Init(in.OutgoingSink, signalClose, wg)
	for i, pipe := range in.pipelines {
		in.start(i, pipe)
	}
	return stopChan
}

func (in *MultiMetricSource) start(index int, pipe *SamplePipeline) {
	in.StartPipeline(&pipe.SamplePipeline, func(isPassive bool, err error) {
		in.LogFinishedPipeline(isPassive, err, fmt.Sprintf("[%v]: Multi-input pipeline %v", in, index))

		in.stoppedPipelines++
		if in.stoppedPipelines >= len(in.pipelines) {
			in.Stop()
		}
	})
}

func (in *MultiMetricSource) Stop() {
	in.StopPipelines()
}

func (in *MultiMetricSource) String() string {
	return fmt.Sprintf("Multi Input (len %v)", len(in.pipelines))
}

func (in *MultiMetricSource) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, len(in.pipelines))
	for i, source := range in.pipelines {
		res[i] = source
	}
	return res
}
