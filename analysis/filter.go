package analysis

import "github.com/antongulenko/data2go/sample"

type SampleFilter struct {
	AbstractProcessor
	Description   string
	IncludeFilter func(inSample *sample.Sample) bool // Return true if sample should be INcluded
}

func (p *SampleFilter) Sample(inSample *sample.Sample, header *sample.Header) error {
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
