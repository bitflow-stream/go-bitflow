package plotHttp

import (
	"sort"

	"github.com/antongulenko/go-bitflow"
)

var HttpPlotter = &httpProcessor{
	data:  make(map[string][]float64),
	names: make([]string, 0, 10),
}

type httpProcessor struct {
	bitflow.AbstractProcessor

	data  map[string][]float64
	names []string
}

func (p *httpProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	p.logSample(sample, header)
	return p.OutgoingSink.Sample(sample, header)
}

func (p *httpProcessor) logSample(sample *bitflow.Sample, header *bitflow.Header) {
	for i, field := range header.Fields {
		val := float64(sample.Values[i])

		if _, ok := p.data[field]; !ok {
			p.names = append(p.names, field)
			sort.Strings(p.names)
		}

		p.data[field] = append(p.data[field], val)
	}
}

func (p *httpProcessor) metricNames() []string {
	return p.names
}

func (p *httpProcessor) metricData(metric string) []float64 {
	if data, ok := p.data[metric]; ok {
		return data
	} else {
		return []float64{}
	}
}
