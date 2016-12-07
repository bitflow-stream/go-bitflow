package pipeline

import (
	"bytes"
	"fmt"

	"github.com/antongulenko/go-bitflow"
)

type SampleFilter struct {
	bitflow.AbstractProcessor
	Description   fmt.Stringer
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
	if p.Description == nil {
		return "Sample Filter"
	} else {
		return "Sample Filter: " + p.Description.String()
	}
}

type SampleTagFilter struct {
	bitflow.AbstractProcessor
	equal   [][2]string // [0]: tag name, [1]: value
	unequal [][2]string
}

func (p *SampleTagFilter) Equal(tag, value string) {
	p.equal = append(p.equal, [2]string{tag, value})
}

func (p *SampleTagFilter) Unequal(tag, value string) {
	p.unequal = append(p.unequal, [2]string{tag, value})
}

func (p *SampleTagFilter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	for _, eq := range p.equal {
		if sample.Tag(eq[0]) != eq[1] {
			return nil
		}
	}
	for _, eq := range p.unequal {
		if sample.Tag(eq[0]) == eq[1] {
			return nil
		}
	}
	return p.OutgoingSink.Sample(sample, header)
}

func (p *SampleTagFilter) MergeProcessor(otherProcessor bitflow.SampleProcessor) bool {
	if other, ok := otherProcessor.(*SampleTagFilter); !ok {
		return false
	} else {
		p.equal = append(p.equal, other.equal...)
		p.unequal = append(p.unequal, other.unequal...)
		return true
	}
}

func (p *SampleTagFilter) String() string {
	var filters bytes.Buffer
	for _, eq := range p.equal {
		if filters.Len() > 0 {
			filters.WriteString(" && ")
		}
		fmt.Fprintf(&filters, "%v == %v", eq[0], eq[1])
	}
	for _, uneq := range p.unequal {
		if filters.Len() > 0 {
			filters.WriteString(" && ")
		}
		fmt.Fprintf(&filters, "%v != %v", uneq[0], uneq[1])
	}
	return fmt.Sprintf("Sample tag filter: %v", filters.String())
}
