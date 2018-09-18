package steps

import (
	"fmt"
	"sort"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
	log "github.com/sirupsen/logrus"
)

// Can tolerate multiple headers, fills missing data up with default values.
type MultiHeaderMerger struct {
	bitflow.NoopProcessor
	header *bitflow.Header

	metrics map[string][]bitflow.Value
	samples []*bitflow.SampleMetadata
}

func NewMultiHeaderMerger() *MultiHeaderMerger {
	return &MultiHeaderMerger{
		metrics: make(map[string][]bitflow.Value),
	}
}

func RegisterMergeHeaders(b reg.ProcessorRegistry) {
	b.RegisterAnalysis("merge_headers",
		func(p *pipeline.SamplePipeline) {
			p.Add(NewMultiHeaderMerger())
		},
		"Accept any number of changing headers and merge them into one output header when flushing the results")
}

func (p *MultiHeaderMerger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	p.addSample(sample, header)
	return nil
}

func (p *MultiHeaderMerger) addSample(incomingSample *bitflow.Sample, header *bitflow.Header) {
	handledMetrics := make(map[string]bool, len(header.Fields))
	for i, field := range header.Fields {
		metrics, ok := p.metrics[field]
		if !ok {
			metrics = make([]bitflow.Value, len(p.samples)) // Filled up with zeroes
		}
		p.metrics[field] = append(metrics, incomingSample.Values[i])
		handledMetrics[field] = true
	}
	for field := range p.metrics {
		if ok := handledMetrics[field]; !ok {
			p.metrics[field] = append(p.metrics[field], 0) // Filled up with zeroes
		}
	}

	p.samples = append(p.samples, incomingSample.Metadata())
}

func (p *MultiHeaderMerger) Close() {
	defer p.CloseSink()
	defer func() {
		// Allow garbage collection
		p.metrics = nil
		p.samples = nil
	}()
	if len(p.samples) == 0 {
		log.Warnln(p.String(), "has no samples stored")
		return
	}
	log.Println(p, "reconstructing and flushing", len(p.samples), "samples with", len(p.metrics), "metrics")
	outHeader := p.reconstructHeader()
	for index := range p.samples {
		outSample := p.reconstructSample(index, outHeader)
		if err := p.NoopProcessor.Sample(outSample, outHeader); err != nil {
			err = fmt.Errorf("Error flushing reconstructed samples: %v", err)
			log.Errorln(err)
			p.Error(err)
			return
		}
	}
}

func (p *MultiHeaderMerger) OutputSampleSize(sampleSize int) int {
	if len(p.metrics) > sampleSize {
		sampleSize = len(p.metrics)
	}
	return sampleSize
}

func (p *MultiHeaderMerger) reconstructHeader() *bitflow.Header {
	fields := make([]string, 0, len(p.metrics))
	for field := range p.metrics {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return &bitflow.Header{Fields: fields}
}

func (p *MultiHeaderMerger) reconstructSample(num int, header *bitflow.Header) *bitflow.Sample {
	values := make([]bitflow.Value, len(p.metrics))
	for i, field := range header.Fields {
		slice := p.metrics[field]
		if len(slice) != len(p.samples) {
			// Should never happen
			panic(fmt.Sprintf("Have %v values for field %v, should be %v", len(slice), field, len(p.samples)))
		}
		values[i] = slice[num]
	}
	return p.samples[num].NewSample(values)
}

func (p *MultiHeaderMerger) String() string {
	return "MultiHeaderMerger"
}
