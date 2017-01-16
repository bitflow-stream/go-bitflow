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
	bitflow.AbstractProcessor
	checker      bitflow.HeaderChecker
	samples      []*bitflow.Sample
	lastFlushTag string

	Steps    []BatchProcessingStep
	FlushTag string // If set, flush every time this tag changes
}

type BatchProcessingStep interface {
	ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error)
	String() string
}

func (p *BatchProcessor) Add(step BatchProcessingStep) *BatchProcessor {
	p.Steps = append(p.Steps, step)
	return p
}

func (p *BatchProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	oldHeader := p.checker.LastHeader
	flush := p.checker.InitializedHeaderChanged(header)
	if p.FlushTag != "" {
		val := sample.Tag(p.FlushTag)
		if oldHeader != nil {
			flush = flush || val != p.lastFlushTag
		}
		p.lastFlushTag = val
	}
	if flush {
		if err := p.flush(oldHeader); err != nil {
			return err
		}
	}
	p.samples = append(p.samples, sample)
	return nil
}

func (p *BatchProcessor) Close() {
	defer p.CloseSink()
	header := p.checker.LastHeader
	if header == nil {
		log.Warnln(p.String(), "received no samples")
	} else if len(p.samples) > 0 {
		if err := p.flush(header); err != nil {
			p.Error(err)
		}
	}
}

func (p *BatchProcessor) flush(header *bitflow.Header) error {
	samples := p.samples
	p.samples = nil // Allow garbage collection
	if samples, header, err := p.executeSteps(samples, header); err != nil {
		return err
	} else {
		if header == nil {
			return fmt.Errorf("Cannot flush %v samples because nil-header was returned by last batch processing step", len(samples))
		}
		log.Println("Flushing", len(samples), "batched samples with", len(header.Fields), "metrics")
		for _, sample := range samples {
			if err := p.OutgoingSink.Sample(sample, header); err != nil {
				return fmt.Errorf("Error flushing batch: %v", err)
			}
		}
		return nil
	}
}

func (p *BatchProcessor) executeSteps(samples []*bitflow.Sample, header *bitflow.Header) ([]*bitflow.Sample, *bitflow.Header, error) {
	if len(p.Steps) > 0 {
		log.Println("Executing", len(p.Steps), "batch processing step(s)")
		for _, step := range p.Steps {
			log.Println("Executing", step, "on", len(samples), "samples with", len(header.Fields), "metrics")
			var err error
			header, samples, err = step.ProcessBatch(header, samples)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return samples, header, nil
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
	flushed := ""
	if p.FlushTag != "" {
		flushed = " flushed with " + p.FlushTag
	}
	return fmt.Sprintf("BatchProcessor%v %v step%s%v", flushed, len(p.Steps), extra, strings.Join(steps, ", "))
}

func (p *BatchProcessor) MergeProcessor(other bitflow.SampleProcessor) bool {
	if otherBatch, ok := other.(*BatchProcessor); !ok {
		return false
	} else {
		p.Steps = append(p.Steps, otherBatch.Steps...)
		return true
	}
}

// ==================== Simple implementation ====================

type SimpleBatchProcessingStep struct {
	Description string
	Process     func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error)
}

func (s *SimpleBatchProcessingStep) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	if process := s.Process; process == nil {
		return nil, nil, fmt.Errorf("%s: Process function is not set")
	} else {
		return process(header, samples)
	}
}

func (s *SimpleBatchProcessingStep) String() string {
	if s.Description == "" {
		return "SimpleBatchProcessingStep"
	} else {
		return s.Description
	}
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
func NewSampleShuffler() *SimpleBatchProcessingStep {
	return &SimpleBatchProcessingStep{
		Description: "sample shuffler",
		Process: func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
			log.Println("Shuffling", len(samples), "samples")
			for i := range samples {
				j := rand.Intn(i + 1)
				samples[i], samples[j] = samples[j], samples[i]
			}
			return header, samples, nil
		},
	}
}

// ==================== Multi-header merger ====================
// Can tolerate multiple headers, fills missing data up with default values.
type MultiHeaderMerger struct {
	bitflow.AbstractProcessor
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
			panic(fmt.Sprintf("Have %v values for field %v, should be %v", len(slice), field, len(p.samples)))
		}
		values[i] = slice[num]
	}
	return p.samples[num].NewSample(values)
}

func (p *MultiHeaderMerger) String() string {
	return "MultiHeaderMerger"
}
