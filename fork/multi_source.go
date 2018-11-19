package fork

import (
	"fmt"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
)

type MultiMetricSource struct {
	MultiPipeline
	bitflow.AbstractSampleProcessor

	pipelines        []*pipeline.SamplePipeline
	stoppedPipelines int
}

func (in *MultiMetricSource) Add(subPipeline *pipeline.SamplePipeline) {
	in.pipelines = append(in.pipelines, subPipeline)
}

func (in *MultiMetricSource) AddSource(source bitflow.SampleSource, steps ...bitflow.SampleProcessor) {
	pipe := &pipeline.SamplePipeline{
		SamplePipeline: bitflow.SamplePipeline{
			Source: source,
		},
	}
	for _, step := range steps {
		pipe.Add(step)
	}
	in.Add(pipe)
}

func (in *MultiMetricSource) Start(wg *sync.WaitGroup) golib.StopChan {
	stopChan := golib.NewStopChan()
	signalClose := func() {
		in.CloseSinkParallel(wg)
		stopChan.Stop()
	}

	in.MultiPipeline.Init(in.GetSink(), signalClose, wg)
	for i, pipe := range in.pipelines {
		in.start(i, pipe)
	}
	return stopChan
}

func (in *MultiMetricSource) start(index int, pipe *pipeline.SamplePipeline) {
	in.StartPipeline(&pipe.SamplePipeline, func(isPassive bool, err error) {
		in.LogFinishedPipeline(isPassive, err, fmt.Sprintf("[%v]: Multi-input pipeline %v", in, index))

		in.stoppedPipelines++
		if in.stoppedPipelines >= len(in.pipelines) {
			in.Close()
		}
	})
}

func (in *MultiMetricSource) Close() {
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
