package pipeline

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/data2go"
	"github.com/antongulenko/golib"
)

// ==================== Abstract Processor ====================
// This type defines standard implementations for all methods in
// SampleProcessor. Parts can be shadowed by embedding the type.
type AbstractProcessor struct {
	data2go.AbstractMetricSource
	data2go.AbstractMetricSink
	stopChan chan error
}

func (p *AbstractProcessor) Header(header *data2go.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		return p.OutgoingSink.Header(header)
	}
}

func (p *AbstractProcessor) Sample(sample *data2go.Sample, header *data2go.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	return p.OutgoingSink.Sample(sample, header)
}

func (p *AbstractProcessor) Check(sample *data2go.Sample, header *data2go.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	return nil
}

func (p *AbstractProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	// Chan buffer of 2 makes sure the Processor can report an error and then
	// call AbstractProcessor.CloseSink() without worrying if the error has been reported previously or not.
	p.stopChan = make(chan error, 2)
	return p.stopChan
}

func (p *AbstractProcessor) CloseSink(wg *sync.WaitGroup) {
	if c := p.stopChan; c != nil {
		// If there was no error, make sure the channel still returns something to signal that this task is done.
		c <- nil
		p.stopChan = nil
	}
	p.AbstractMetricSource.CloseSink(wg)
}

func (p *AbstractProcessor) Error(err error) {
	if c := p.stopChan; c != nil {
		c <- err
	}
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
	Sample *data2go.Sample
	Header *data2go.Header
}

func (p *DecouplingProcessor) Header(header *data2go.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		p.samples <- TaggedSample{Header: header}
		return nil
	}
}

func (p *DecouplingProcessor) Sample(sample *data2go.Sample, header *data2go.Header) error {
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
	AbstractProcessor
}

func (p *SamplePrinter) Header(header *data2go.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		log.Println("Processing Header len:", len(header.Fields), "tags:", header.HasTags)
		return p.OutgoingSink.Header(header)
	}
}

func (p *SamplePrinter) Sample(sample *data2go.Sample, header *data2go.Header) error {
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
