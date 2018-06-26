package pipeline

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
	log "github.com/sirupsen/logrus"
)

type BatchProcessor struct {
	bitflow.NoopProcessor
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

type ResizingBatchProcessingStep interface {
	BatchProcessingStep
	OutputSampleSize(sampleSize int) int
}

func (p *BatchProcessor) OutputSampleSize(sampleSize int) int {
	for _, step := range p.Steps {
		if step, ok := step.(ResizingBatchProcessingStep); ok {
			sampleSize = step.OutputSampleSize(sampleSize)
		}
	}
	return sampleSize
}

func (p *BatchProcessor) Add(step BatchProcessingStep) *BatchProcessor {
	p.Steps = append(p.Steps, step)
	return p
}

func (p *BatchProcessor) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, len(p.Steps))
	for i, step := range p.Steps {
		res[i] = step
	}
	return res
}

func (p *BatchProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
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
			if err := p.NoopProcessor.Sample(sample, header); err != nil {
				return fmt.Errorf("Error flushing batch: %v", err)
			}
		}
		return nil
	}
}

func (p *BatchProcessor) executeSteps(samples []*bitflow.Sample, header *bitflow.Header) ([]*bitflow.Sample, *bitflow.Header, error) {
	if len(p.Steps) > 0 {
		log.Println("Executing", len(p.Steps), "batch processing step(s)")
		for i, step := range p.Steps {
			if len(samples) == 0 {
				log.Warnln("Cannot execute remaining", len(p.Steps)-i, "batch step(s) because the batch with", len(header.Fields), "has no samples")
				break
			} else {
				log.Println("Executing", step, "on", len(samples), "samples with", len(header.Fields), "metrics")
				var err error
				header, samples, err = step.ProcessBatch(header, samples)
				if err != nil {
					return nil, nil, err
				}
			}
		}
	}
	return samples, header, nil
}

func (p *BatchProcessor) String() string {
	extra := "s"
	if len(p.Steps) == 1 {
		extra = ""
	}
	flushed := ""
	if p.FlushTag != "" {
		flushed = ", flushed with " + p.FlushTag
	}
	return fmt.Sprintf("BatchProcessor (%v step%s%s)", len(p.Steps), extra, flushed)
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
	Description          string
	Process              func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error)
	OutputSampleSizeFunc func(sampleSize int) int
}

func (s *SimpleBatchProcessingStep) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	if process := s.Process; process == nil {
		return nil, nil, fmt.Errorf("%v: Process function is not set", s)
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

func (s *SimpleBatchProcessingStep) OutputSampleSize(sampleSize int) int {
	if f := s.OutputSampleSizeFunc; f != nil {
		return f(sampleSize)
	}
	return sampleSize
}
