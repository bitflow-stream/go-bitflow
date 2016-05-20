package analysis

import (
	"fmt"
	"log"
	"sync"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

// ==================== Abstract Processor ====================
// This type defines standard implementations for all methods in
// SampleProcessor. Parts can be shadowed by embedding the type.
type AbstractProcessor struct {
	sample.AbstractMetricSource
	sample.AbstractMetricSink
}

func (p *AbstractProcessor) Header(header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		return p.OutgoingSink.Header(header)
	}
}

func (p *AbstractProcessor) Sample(sample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	return p.OutgoingSink.Sample(sample, header)
}

func (p *AbstractProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	return nil
}

func (p *AbstractProcessor) Close() {
	// Propagate the Close() invocation
	p.CloseSink(nil)
}

func (p *AbstractProcessor) String() string {
	return "AbstractProcessor"
}

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
		return p.OutgoingSink.Header(*sample.Header)
	} else {
		return p.OutgoingSink.Sample(*sample.Sample, *sample.Header)
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
