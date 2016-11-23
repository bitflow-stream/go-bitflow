package pipeline

import "github.com/antongulenko/go-bitflow"

type MetricStripper struct {
	bitflow.AbstractProcessor
	header     *bitflow.Header
	headerSent bool
}

func NewMetricStripper() *MetricStripper {
	return &MetricStripper{
		header: &bitflow.Header{
			HasTags: true,
		},
	}
}

func (p *MetricStripper) Header(header *bitflow.Header) (err error) {
	if !p.headerSent {
		err = p.AbstractProcessor.Header(p.header)
		p.headerSent = err == nil
	}
	return
}

func (p *MetricStripper) Sample(sample *bitflow.Sample, _ *bitflow.Header) (err error) {
	sample = sample.Metadata().NewSample(nil)
	err = p.AbstractProcessor.Sample(sample, p.header)
	if err != nil {
		p.headerSent = false
	}
	return
}
