package pipeline

import "github.com/antongulenko/go-bitflow"

type SampleFilter struct {
	AbstractProcessor
	Description   string
	IncludeFilter func(inSample *bitflow.Sample) bool // Return true if sample should be INcluded
}

func (p *SampleFilter) Sample(inSample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(inSample, header); err != nil {
		return err
	}
	if filter := p.IncludeFilter; filter != nil && filter(inSample) {
		return p.OutgoingSink.Sample(inSample, header)
	} else {
		return nil
	}
}

func (p *SampleFilter) String() string {
	if p.Description == "" {
		return "Sample Filter"
	} else {
		return "Sample Filter: " + p.Description
	}
}
