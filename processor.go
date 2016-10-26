package pipeline

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

// ==================== DecouplingProcessor ====================
// Decouple the incoming samples from the MetricSink through a
// looping goroutine and a channel. Creates potential parallelism in the pipeline.
type DecouplingProcessor struct {
	bitflow.AbstractProcessor
	samples       chan TaggedSample
	loopTask      *golib.LoopTask
	ChannelBuffer int // Must be set before calling Start()
}

type TaggedSample struct {
	Sample *bitflow.Sample
	Header *bitflow.Header
}

func (p *DecouplingProcessor) Header(header *bitflow.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		p.samples <- TaggedSample{Header: header}
		return nil
	}
}

func (p *DecouplingProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	p.samples <- TaggedSample{Sample: sample, Header: header}
	return nil
}

func (p *DecouplingProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.samples = make(chan TaggedSample, p.ChannelBuffer)
	p.loopTask = golib.NewErrLoopTask(p.String(), func(stop golib.StopChan) error {
		select {
		case sample, open := <-p.samples:
			if open {
				if err := p.forward(sample); err != nil {
					return fmt.Errorf("Error forwarding sample from %v to %v: %v", p, p.OutgoingSink, err)
				}
			} else {
				p.loopTask.EnableOnly()
			}
		case <-stop:
		}
		return nil
	})
	p.loopTask.StopHook = func() {
		p.CloseSink(wg)
	}
	return p.loopTask.Start(wg)
}

func (p *DecouplingProcessor) forward(sample TaggedSample) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if sample.Sample == nil {
		return p.OutgoingSink.Header(sample.Header)
	} else {
		return p.OutgoingSink.Sample(sample.Sample, sample.Header)
	}
}

func (p *DecouplingProcessor) Close() {
	close(p.samples)
}

func (p *DecouplingProcessor) String() string {
	return "DecouplingProcessor"
}

// ==================== SamplePrinter (example) ====================
type SamplePrinter struct {
	bitflow.AbstractProcessor
}

func (p *SamplePrinter) Header(header *bitflow.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		log.Println("Processing Header len:", len(header.Fields), "tags:", header.HasTags)
		return p.OutgoingSink.Header(header)
	}
}

func (p *SamplePrinter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	tags := ""
	if sample.NumTags() > 0 {
		tags = "(" + sample.TagString() + ")"
	}
	log.Println("Processing Sample time:", sample.Time, "len:", len(sample.Values), tags)
	return p.OutgoingSink.Sample(sample, header)
}

func (p *SamplePrinter) String() string {
	return "SamplePrinter"
}
