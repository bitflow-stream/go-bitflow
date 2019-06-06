package steps

import (
	"fmt"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

// Decouple the incoming samples from the MetricSink through a
// looping goroutine and a channel. Creates potential parallelism in the pipeline.
type DecouplingProcessor struct {
	bitflow.NoopProcessor
	samples       chan bitflow.SampleAndHeader
	loopTask      *golib.LoopTask
	ChannelBuffer int // Must be set before calling Start()
}

func AddDecoupleStep(p *bitflow.SamplePipeline, params map[string]interface{}) error {
	p.Add(&DecouplingProcessor{ChannelBuffer: params["buf"].(int)})
	return nil
}

func RegisterDecouple(b reg.ProcessorRegistry) {
	b.RegisterStep("decouple", AddDecoupleStep,
		"Start a new concurrent routine for handling samples. The parameter is the size of the FIFO-buffer for handing over the samples").
		Required("buf", reg.Int())
}

func (p *DecouplingProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	p.samples <- bitflow.SampleAndHeader{Sample: sample, Header: header}
	return nil
}

func (p *DecouplingProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.samples = make(chan bitflow.SampleAndHeader, p.ChannelBuffer)
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

func (p *DecouplingProcessor) forward(sample bitflow.SampleAndHeader) error {
	return p.NoopProcessor.Sample(sample.Sample, sample.Header)
}

func (p *DecouplingProcessor) Close() {
	close(p.samples)
}

func (p *DecouplingProcessor) String() string {
	return fmt.Sprintf("DecouplingProcessor (buffer %v)", p.ChannelBuffer)
}
