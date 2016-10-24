package pipeline

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
)

type BatchProcessor struct {
	AbstractProcessor
	header  *bitflow.Header
	samples []*bitflow.Sample

	Steps []BatchProcessingStep
}

type BatchProcessingStep interface {
	// TODO Optimize implementors of BatchProcessingStep to reuse the sample-array instead of allocating a new one with the same size.
	ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error)
	String() string
}

func (p *BatchProcessor) Add(step BatchProcessingStep) *BatchProcessor {
	p.Steps = append(p.Steps, step)
	return p
}

func (p *BatchProcessor) checkHeader(header *bitflow.Header) error {
	if !p.header.Equals(header) {
		return fmt.Errorf("%v does not allow changing headers", p)
	}
	return nil
}

func (p *BatchProcessor) Header(header *bitflow.Header) error {
	if p.header == nil {
		p.header = header
	} else if err := p.checkHeader(header); err != nil {
		return err
	}
	return p.CheckSink()
}

func (p *BatchProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	if err := p.checkHeader(header); err != nil {
		return err
	}
	p.samples = append(p.samples, sample)
	return nil
}

func (p *BatchProcessor) Close() {
	defer p.CloseSink(nil)
	defer func() {
		// Allow garbage collection
		p.samples = nil
	}()
	if p.header == nil {
		log.Warnln(p.String(), "has no samples stored")
		return
	}
	if err := p.executeSteps(); err != nil {
		p.Error(err)
	} else {
		log.Println("Flushing", len(p.samples), "batched samples")
		if err := p.OutgoingSink.Header(p.header); err != nil {
			p.Error(fmt.Errorf("Error flushing batch header: %v", err))
			return
		}
		for _, sample := range p.samples {
			if err := p.OutgoingSink.Sample(sample, p.header); err != nil {
				p.Error(fmt.Errorf("Error flushing batch: %v", err))
				return
			}
		}
	}
}

func (p *BatchProcessor) executeSteps() error {
	if len(p.Steps) > 0 {
		log.Println("Executing", len(p.Steps), "batch processing step(s)")
		for _, step := range p.Steps {
			log.Println("Executing", step, "on", len(p.samples), "samples with", len(p.header.Fields), "metrics")
			var err error
			p.header, p.samples, err = step.ProcessBatch(p.header, p.samples)
			if err != nil {
				return err
			}
		}
		if p.header == nil {
			return fmt.Errorf("Cannot flush %v samples because no valid header was returned by last batch processing step", len(p.samples))
		}
	}
	return nil
}

func (p *BatchProcessor) String() string {
	steps := make([]string, len(p.Steps))
	for i, step := range p.Steps {
		steps[i] = step.String()
	}
	var extra string
	if len(p.Steps) == 0 {
		extra = "s"
	} else if len(p.Steps) == 1 {
		extra = ": "
	} else {
		extra = "s: "
	}
	return fmt.Sprintf("BatchProcessor %v step%s%v", len(p.Steps), extra, strings.Join(steps, ", "))
}

// ==================== Tag & Timestamp sort ====================
// Sort based on given Tags, use Timestamp as last sort criterion
type SampleSorter struct {
	Tags []string
}

type SampleSlice struct {
	samples []*bitflow.Sample
	sorter  *SampleSorter
}

func (s SampleSlice) Len() int {
	return len(s.samples)
}

func (s SampleSlice) Less(i, j int) bool {
	a := s.samples[i]
	b := s.samples[j]
	for _, tag := range s.sorter.Tags {
		tagA := a.Tag(tag)
		tagB := b.Tag(tag)
		if tagA == tagB {
			continue
		}
		return tagA < tagB
	}
	return a.Time.Before(b.Time)
}

func (s SampleSlice) Swap(i, j int) {
	s.samples[i], s.samples[j] = s.samples[j], s.samples[i]
}

func (sorter *SampleSorter) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	log.Println("Sorting", len(samples), "samples")
	sort.Sort(SampleSlice{samples, sorter})
	return header, samples, nil
}

func (sorter *SampleSorter) String() string {
	all := make([]string, len(sorter.Tags)+1)
	copy(all, sorter.Tags)
	all[len(all)-1] = "Timestamp"
	return "Sort: " + strings.Join(all, ", ")
}

// ==================== Sample Shuffler ====================
type SampleShuffler struct {
}

func (*SampleShuffler) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	log.Println("Shuffling", len(samples), "samples")
	for i := range samples {
		j := rand.Intn(i + 1)
		samples[i], samples[j] = samples[j], samples[i]
	}
	return header, samples, nil
}

func (*SampleShuffler) String() string {
	return "SampleShuffler"
}

// ==================== Multi-header merger ====================
// Can tolerate multiple headers, fills missing data up with default values.
type MultiHeaderMerger struct {
	AbstractProcessor
	header *bitflow.Header

	hasTags bool
	metrics map[string][]bitflow.Value
	samples []*bitflow.SampleMetadata
}

func NewMultiHeaderMerger() *MultiHeaderMerger {
	return &MultiHeaderMerger{
		metrics: make(map[string][]bitflow.Value),
	}
}

func (p *MultiHeaderMerger) Header(_ *bitflow.Header) error {
	return p.CheckSink()
}

func (p *MultiHeaderMerger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
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
	p.hasTags = p.hasTags || header.HasTags
}

func (p *MultiHeaderMerger) Close() {
	defer p.CloseSink(nil)
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
	if err := p.OutgoingSink.Header(outHeader); err != nil {
		err = fmt.Errorf("Error flushing reconstructed header: %v", err)
		log.Errorln(err)
		p.Error(err)
		return
	}
	for index := range p.samples {
		outSample := p.reconstructSample(index, outHeader)
		if err := p.OutgoingSink.Sample(outSample, outHeader); err != nil {
			err = fmt.Errorf("Error flushing reconstructed samples: %v", err)
			log.Errorln(err)
			p.Error(err)
			return
		}
	}
}

func (p *MultiHeaderMerger) reconstructHeader() *bitflow.Header {
	fields := make([]string, 0, len(p.metrics))
	for field := range p.metrics {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return &bitflow.Header{Fields: fields, HasTags: p.hasTags}
}

func (p *MultiHeaderMerger) reconstructSample(num int, header *bitflow.Header) *bitflow.Sample {
	values := make([]bitflow.Value, len(p.metrics))
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
