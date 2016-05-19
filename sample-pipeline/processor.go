package pipeline

import (
	"log"
	"sync"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

// ==================== DecouplingProcessor ====================
// Decouple the incoming samples from the MetricSink through a
// looping goroutine and a channel. Creates potential parallelism in the pipeline.
type DecouplingProcessor struct {
	AbstractProcessor
	samples       chan TaggedSample
	loopTask      *golib.LoopTask
	ChannelBuffer int // Must be set before calling Start()
}

type TaggedSample struct {
	Sample *sample.Sample
	Header *sample.Header
}

func (p *DecouplingProcessor) Header(header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		p.samples <- TaggedSample{Header: &header}
		return nil
	}
}

func (p *DecouplingProcessor) Sample(sample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	p.samples <- TaggedSample{Sample: &sample, Header: &header}
	return nil
}

func (p *DecouplingProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.samples = make(chan TaggedSample, p.ChannelBuffer)
	p.loopTask = golib.NewLoopTask(p.String(), func(stop golib.StopChan) {
		select {
		case sample, open := <-p.samples:
			if open {
				if err := p.forward(sample); err != nil {
					log.Printf("Error forwarding sample to from %v to %v: %v\n", p, p.OutgoingSink, err)
				}
			} else {
				p.loopTask.EnableOnly()
			}
		case <-stop:
		}
	})
	p.loopTask.StopHook = p.loopStopped
	return p.loopTask.Start(wg)
}

func (p *DecouplingProcessor) forward(sample TaggedSample) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if sample.Sample == nil {
		return p.OutgoingSink.Header(*sample.Header)
	} else {
		return p.OutgoingSink.Sample(*sample.Sample, *sample.Header)
	}
}

func (p *DecouplingProcessor) Close() {
	close(p.samples)
}

func (p *DecouplingProcessor) loopStopped() {
	p.CloseSink()
}

func (p *DecouplingProcessor) String() string {
	return "DecouplingProcessor"
}

// ==================== SamplePrinter ====================
type SamplePrinter struct {
	AbstractProcessor
}

func (p *SamplePrinter) Header(header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		log.Printf("Processing Header len %v, tags: %v\n", len(header.Fields), header.HasTags)
		return p.OutgoingSink.Header(header)
	}
}

func (p *SamplePrinter) Sample(sample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	tags := ""
	if len(sample.Tags) > 0 {
		tags = " (" + sample.TagString() + ")"
	}
	log.Printf("Processing Sample from %v, len %v%v\n", sample.Time, len(sample.Values), tags)
	return p.OutgoingSink.Sample(sample, header)
}

func (p *SamplePrinter) String() string {
	return "SamplePrinter"
}
