package pipeline

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
)

type SampleFilter struct {
	bitflow.NoopProcessor
	Description   fmt.Stringer
	IncludeFilter func(sample *bitflow.Sample, header *bitflow.Header) (bool, error) // Return true if sample should be included
}

func (p *SampleFilter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	filter := p.IncludeFilter
	if filter != nil {
		res, err := filter(sample, header)
		if err != nil {
			return err
		}
		if res {
			return p.NoopProcessor.Sample(sample, header)
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
