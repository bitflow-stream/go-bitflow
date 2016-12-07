package pipeline

import "github.com/antongulenko/go-bitflow"

var emptyHeader = &bitflow.Header{
	HasTags: true,
}

type MetricStripper struct {
	bitflow.AbstractProcessor
}

func (p *MetricStripper) Sample(sample *bitflow.Sample, _ *bitflow.Header) error {
	sample = sample.Metadata().NewSample(nil)
	return p.AbstractProcessor.Sample(sample, emptyHeader)
}
