package pipeline

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
)

type SampleFilter struct {
	bitflow.AbstractProcessor
	Description   fmt.Stringer
	IncludeFilter func(sample *bitflow.Sample, header *bitflow.Header) (bool, error) // Return true if sample should be included
}

func (p *SampleFilter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	filter := p.IncludeFilter
	if filter != nil {
		res, err := filter(sample, header)
		if err != nil {
			return err
		}
		if res {
			return p.OutgoingSink.Sample(sample, header)
		}
	}
	return nil
}

func (p *SampleFilter) String() string {
	if p.Description == nil {
		return "Sample Filter"
	} else {
		return "Sample Filter: " + p.Description.String()
	}
}
