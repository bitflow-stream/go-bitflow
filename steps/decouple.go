package steps

import (
	"fmt"
	"strconv"
	"sync"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
)

// Decouple the incoming samples from the MetricSink through a
// looping goroutine and a channel. Creates potential parallelism in the pipeline.
type DecouplingProcessor struct {
	bitflow.NoopProcessor
	samples       chan TaggedSample
	loopTask      *golib.LoopTask
	ChannelBuffer int // Must be set before calling Start()
}

func RegisterDecouple(b *query.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) error {
		buf, err := strconv.Atoi(params["batch"])
		if err != nil {
			err = query.ParameterError("batch", err)
		} else {
			p.Add(&DecouplingProcessor{ChannelBuffer: buf})
		}
		return err
	}
	b.RegisterAnalysisParamsErr("decouple", create, "Start a new concurrent routine for handling samples. The parameter is the size of the FIFO-buffer for handing over the samples", []string{"batch"})
}

type TaggedSample struct {
	Sample *bitflow.Sample
	Header *bitflow.Header
}

func (p *DecouplingProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	p.samples <- TaggedSample{Sample: sample, Header: header}
	return nil
}

func (p *DecouplingProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.samples = make(chan TaggedSample, p.ChannelBuffer)
	p.loopTask = &golib.LoopTask{
		Description: p.String(),
		StopHook:    p.CloseSink,
		Loop: func(stop golib.StopChan) error {
			select {
			case sample, open := <-p.samples:
				if open {
					if err := p.forward(sample); err != nil {
						return fmt.Errorf("Error forwarding sample from %v to %v: %v", p, p.GetSink(), err)
					}
				} else {
					p.loopTask.Stop()
				}
			case <-stop.WaitChan():
			}
			return nil
		},
	}
	return p.loopTask.Start(wg)
}

func (p *DecouplingProcessor) forward(sample TaggedSample) error {
	return p.NoopProcessor.Sample(sample.Sample, sample.Header)
}

func (p *DecouplingProcessor) Close() {
	close(p.samples)
}

func (p *DecouplingProcessor) String() string {
	return fmt.Sprintf("DecouplingProcessor (buffer %v)", p.ChannelBuffer)
}
