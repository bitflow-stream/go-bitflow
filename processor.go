package pipeline

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

// MergableProcessor is an extension of bitflow.SampleProcessor, that also allows
// merging two processor instances of the same time into one. Merging is only allowed
// when the result of the merge would has exactly the same functionality as using the
// two separate instances. This can be used as an optional optimization.
type MergableProcessor interface {
	bitflow.SampleProcessor
	MergeProcessor(other bitflow.SampleProcessor) bool
}

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
		p.CloseSink()
	}
	return p.loopTask.Start(wg)
}

func (p *DecouplingProcessor) forward(sample TaggedSample) error {
	return p.OutgoingSink.Sample(sample.Sample, sample.Header)
}

func (p *DecouplingProcessor) Close() {
	close(p.samples)
}

func (p *DecouplingProcessor) String() string {
	return fmt.Sprintf("DecouplingProcessor (buffer %v)", p.ChannelBuffer)
}

// ==================== SimpleProcessor ====================

type SimpleProcessor struct {
	bitflow.AbstractProcessor
	Description string
	Process     func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error)
	OnClose     func()
}

func (p *SimpleProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	if process := p.Process; process == nil {
		return fmt.Errorf("%s: Process function is not set", p)
	} else {
		sample, header, err := process(sample, header)
		if err == nil && sample != nil && header != nil {
			err = p.OutgoingSink.Sample(sample, header)
		}
		return err
	}
}

func (p *SimpleProcessor) Close() {
	if c := p.OnClose; c != nil {
		c()
	}
	p.AbstractProcessor.Close()
}

func (p *SimpleProcessor) String() string {
	if p.Description == "" {
		return "SimpleProcessor"
	} else {
		return p.Description
	}
}

// ==================== SamplePrinter (example) ====================

func NewSamplePrinter() *SimpleProcessor {
	var checker bitflow.HeaderChecker
	return &SimpleProcessor{
		Description: "sample printer",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			if checker.HeaderChanged(header) {
				log.Println("Processing Header len:", len(header.Fields), "tags:", header.HasTags)
			}
			tags := ""
			if sample.NumTags() > 0 {
				tags = "(" + sample.TagString() + ")"
			}
			log.Println("Processing Sample time:", sample.Time, "len:", len(sample.Values), tags)
			return sample, header, nil
		},
	}
}
