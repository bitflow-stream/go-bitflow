package pipeline

import (
	"log"

	"github.com/antongulenko/data2go/sample"
)

// ==================== SamplePrinter ====================
type SamplePrinter struct {
	AbstractProcessor
}

func (p *SamplePrinter) Header(header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else {
		log.Println("Processing Header:", header)
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
	log.Println("Processing Sample:", sample)
	return p.OutgoingSink.Sample(sample, header)
}

func (p *SamplePrinter) String() string {
	return "SamplePrinter"
}
