package pipeline

import (
	"fmt"
	"log"

	"github.com/antongulenko/data2go/sample"
)

type BatchProcessor struct {
	AbstractProcessor
	header  *sample.Header
	samples []*sample.Sample
}

func (p *BatchProcessor) checkHeader(header *sample.Header) error {
	if !p.header.Equals(header) {
		return fmt.Errorf("%v does not allow changing headers\n", p)
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
	if p.header == nil {
		log.Println(p.String() + " has no samples stored")
		return
	}
	log.Println("Now flushing", len(p.samples), "batched samples")
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
	p.CloseSink()
}

func (p *BatchProcessor) String() string {
	return "BatchProcessor"
}
