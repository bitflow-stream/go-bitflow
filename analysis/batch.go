package analysis

import (
	"fmt"
	"log"
	"sort"

	"github.com/antongulenko/data2go/sample"
)

type BatchProcessor struct {
	AbstractProcessor
	header  *sample.Header
	samples []*sample.Sample

	Steps []BatchProcessingStep
}

type BatchProcessingStep interface {
	ProcessBatch(header *sample.Header, samples []*sample.Sample) (*sample.Header, []*sample.Sample)
	String() string
}

func (p *BatchProcessor) Add(step BatchProcessingStep) *BatchProcessor {
	p.Steps = append(p.Steps, step)
	return p
}

func (p *BatchProcessor) checkHeader(header *sample.Header) error {
	if !p.header.Equals(header) {
		return fmt.Errorf("%v does not allow changing headers", p)
	}
	return nil
}

func (p *BatchProcessor) Header(header sample.Header) error {
	if p.header == nil {
		p.header = &header
	} else if err := p.checkHeader(&header); err != nil {
		return err
	}
	return p.CheckSink()
}

func (p *BatchProcessor) Sample(sample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	if err := p.checkHeader(&header); err != nil {
		return err
	}
	p.samples = append(p.samples, &sample)
	return nil
}

func (p *BatchProcessor) Close() {
	defer p.CloseSink(nil)
	if p.header == nil {
		log.Println(p.String(), "has no samples stored")
		return
	}
	p.executeSteps()
	log.Println("Flushing", len(p.samples), "batched samples")
	if err := p.OutgoingSink.Header(*p.header); err != nil {
		log.Println("Error flushing batch header:", err)
		return
	}
	for _, sample := range p.samples {
		if err := p.OutgoingSink.Sample(*sample, *p.header); err != nil {
			log.Println("Error flushing batch:", err)
			return
		}
	}
}

func (p *BatchProcessor) executeSteps() {
	if len(p.Steps) > 0 {
		log.Printf("Executing %v batch processing step(s)...\n", len(p.Steps))
		header := p.header
		samples := p.samples
		for _, step := range p.Steps {
			log.Printf("Executing %v on %v samples with %v metrics...\n", step, len(samples), len(header.Fields))
			header, samples = step.ProcessBatch(header, samples)
		}
		p.header = header
		p.samples = samples
	}
}

func (p *BatchProcessor) String() string {
	return "BatchProcessor"
}

// ==================== Timestamp sort ====================
type TimestampSort struct {
}

type SampleSlice []*sample.Sample

func (s SampleSlice) Len() int {
	return len(s)
}

func (s SampleSlice) Less(i, j int) bool {
	return s[i].Time.Before(s[j].Time)
}

func (s SampleSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (*TimestampSort) ProcessBatch(header *sample.Header, samples []*sample.Sample) (*sample.Header, []*sample.Sample) {
	// TODO this sort is not interruptible...
	log.Printf("Sorting %v samples...\n", len(samples))
	sort.Sort(SampleSlice(samples))
	return header, samples
}

func (*TimestampSort) String() string {
	return "TimestampSort"
}

// ==================== Multi-header merger ====================
// Can tolerate multiple headers, fills missing data up with default values.
type MultiHeaderMerger struct {
	AbstractProcessor
	header *sample.Header

	hasTags bool
	metrics map[string][]sample.Value
	samples []sample.SampleMetadata
}

func NewMultiHeaderMerger() *MultiHeaderMerger {
	return &MultiHeaderMerger{
		metrics: make(map[string][]sample.Value),
	}
}

func (p *MultiHeaderMerger) Header(_ sample.Header) error {
	return p.CheckSink()
}

func (p *MultiHeaderMerger) Sample(sample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	p.addSample(sample, header)
	return nil
}

func (p *MultiHeaderMerger) addSample(incomingSample sample.Sample, header sample.Header) {
	handledMetrics := make(map[string]bool, len(header.Fields))
	for i, field := range header.Fields {
		metrics, ok := p.metrics[field]
		if !ok {
			metrics = make([]sample.Value, len(p.samples)) // Filled up with zeroes
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
	p.hasTags = p.hasTags || header.HasTags
}

func (p *MultiHeaderMerger) Close() {
	defer p.CloseSink(nil)
	if len(p.samples) == 0 {
		log.Println(p.String(), "has no samples stored")
		return
	}
	log.Println(p, "reconstructing and flushing", len(p.samples), "samples with", len(p.metrics), "metrics")
	outHeader := p.reconstructHeader()
	if err := p.OutgoingSink.Header(outHeader); err != nil {
		log.Println("Error flushing reconstructed header:", err)
		return
	}
	for index := range p.samples {
		outSample := p.reconstructSample(index, outHeader)
		if err := p.OutgoingSink.Sample(outSample, outHeader); err != nil {
			log.Println("Error flushing reconstructed samples:", err)
			return
		}
	}
}

func (p *MultiHeaderMerger) reconstructHeader() sample.Header {
	fields := make([]string, 0, len(p.metrics))
	for field := range p.metrics {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return sample.Header{Fields: fields, HasTags: p.hasTags}
}

func (p *MultiHeaderMerger) reconstructSample(num int, header sample.Header) sample.Sample {
	values := make([]sample.Value, len(p.metrics))
	for i, field := range header.Fields {
		slice := p.metrics[field]
		if len(slice) != len(p.samples) {
			// Should never happen
			panic(fmt.Sprintf("Only %v values for field %v, should be %v", len(slice), field, len(p.samples)))
		}
		values[i] = slice[num]
	}
	return p.samples[num].NewSample(values)
}

func (p *MultiHeaderMerger) String() string {
	return "MultiHeaderMerger"
}
