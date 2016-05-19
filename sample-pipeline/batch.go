package pipeline

import (
	"fmt"
	"log"
	"sort"

	"github.com/antongulenko/data2go/sample"
)

type BatchProcessor struct {
	AbstractProcessor
	header  *sample.Header
	samples []*sample.Sample

	Steps []BatchProcessingStep
}

type BatchProcessingStep interface {
	ProcessBatch(header *sample.Header, samples []*sample.Sample) (*sample.Header, []*sample.Sample)
	String() string
}

func (p *BatchProcessor) Add(step BatchProcessingStep) *BatchProcessor {
	p.Steps = append(p.Steps, step)
	return p
}

func (p *BatchProcessor) checkHeader(header *sample.Header) error {
	if !p.header.Equals(header) {
		return fmt.Errorf("%v does not allow changing headers", p)
	}
	return nil
}

func (p *BatchProcessor) Header(header sample.Header) error {
	if p.header == nil {
		p.header = &header
	} else if err := p.checkHeader(&header); err != nil {
		return err
	}
	return p.CheckSink()
}

func (p *BatchProcessor) Sample(sample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	if err := p.checkHeader(&header); err != nil {
		return err
	}
	p.samples = append(p.samples, &sample)
	return nil
}

func (p *BatchProcessor) Close() {
	defer p.CloseSink()
	if p.header == nil {
		log.Println(p.String(), "has no samples stored")
		return
	}
	p.executeSteps()
	log.Println("Flushing", len(p.samples), "batched samples")
	if err := p.OutgoingSink.Header(*p.header); err != nil {
		log.Println("Error flushing batch header:", err)
		return
	}
	for _, sample := range p.samples {
		if err := p.OutgoingSink.Sample(*sample, *p.header); err != nil {
			log.Println("Error flushing batch:", err)
			return
		}
	}
}

func (p *BatchProcessor) executeSteps() {
	if len(p.Steps) > 0 {
		log.Printf("Executing %v batch processing steps...\n", len(p.Steps))
		header := p.header
		samples := p.samples
		for _, step := range p.Steps {
			log.Printf("Executing %v on %v samples with %v metrics...\n", step, len(samples), len(header.Fields))
			header, samples = step.ProcessBatch(header, samples)
		}
	}
}

func (p *BatchProcessor) String() string {
	return "BatchProcessor"
}

// ==================== Timestamp sort ====================
type TimestampSort struct {
}

type SampleSlice []*sample.Sample

func (s SampleSlice) Len() int {
	return len(s)
}

func (s SampleSlice) Less(i, j int) bool {
	return s[i].Time.Before(s[j].Time)
}

func (s SampleSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (*TimestampSort) ProcessBatch(header *sample.Header, samples []*sample.Sample) (*sample.Header, []*sample.Sample) {
	// TODO this sort is not interruptible...
	log.Printf("Sorting %v samples...\n", len(samples))
	sort.Sort(SampleSlice(samples))
	return header, samples
}

func (*TimestampSort) String() string {
	return "TimestampSort"
}
