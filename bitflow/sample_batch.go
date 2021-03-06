package bitflow

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

type BatchProcessor struct {
	NoopProcessor
	checker  HeaderChecker
	samples  []*Sample
	shutdown bool

	Steps []BatchProcessingStep

	// ForwardImmediately enables a mode where each received sample is immediately forwarded to the subsequent processing
	// step for further processing. The sample is also stored for batch processing, and all flush semantics are still active,
	// but the results of the batch processing will be dropped. Errors that occur during batch flushing are logged, but are
	// not reported to the caller of the Sample() method.
	ForwardImmediately bool

	// DontFlushOnClose disables the automatic final flushing when this processing step is closed. This is useful only in
	// certain special cases.
	DontFlushOnClose bool

	FlushNoSampleTimeout time.Duration // If > 0, flush when no new samples are received for the given duration. The wall-time is used for this (not sample timestamps)
	FlushSampleLag       time.Duration // If > 0, flush when a sample is received with a timestamp jump bigger than this
	FlushAfterNumSamples int           // If > 0, flush after batch  window contains this amount of samples
	FlushAfterTime       time.Duration // If > 0, flush after time difference between the first and the last received sample in batch is greater than this
	lastAutoFlushError   error
	lastSample           time.Time // Wall time when receiving last sample
	lastSampleTimestamp  time.Time // Timestamp of last sample

	FlushTags     []string // If set, flush every time any of these tags change
	lastFlushTags []string
	flushHeader   *Header
	flushTrigger  *golib.TimeoutCond // Used to trigger flush and to notify about finished flush. Relies on Sample()/Close() being synchronized externally.
	flushError    error
}

type BatchProcessingStep interface {
	ProcessBatch(header *Header, samples []*Sample) (*Header, []*Sample, error)
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

func (p *BatchProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.flushTrigger = golib.NewTimeoutCond(new(sync.Mutex))
	wg.Add(1)
	go p.loopFlush(wg)
	return p.NoopProcessor.Start(wg)
}

func (p *BatchProcessor) Sample(sample *Sample, header *Header) (err error) {
	oldHeader := p.checker.LastHeader
	flush := p.checker.InitializedHeaderChanged(header)
	if p.FlushAfterNumSamples > 0 { // FlushAfterNumSamples and  FlushAfterTime are mutually exclusive. FlushAfterNumSamples is prioritized over FlushAfterTime.
		flush = len(p.samples) >= p.FlushAfterNumSamples
	} else if p.FlushAfterTime > 0 {
		if len(p.samples) > 1 { // At least 2 samples are required
			start := p.samples[0].Time
			end := p.samples[len(p.samples)-1].Time
			diff := end.Sub(start)
			flush = (p.FlushAfterTime - diff) <= 0
		}
	}
	if len(p.FlushTags) > 0 {
		values := make([]string, len(p.FlushTags))
		for i, tag := range p.FlushTags {
			values[i] = sample.Tag(tag)
			if oldHeader != nil && len(p.lastFlushTags) > i && p.lastFlushTags[i] != values[i] {
				flush = true
			}
		}
		p.lastFlushTags = values
	}
	if p.FlushSampleLag > 0 {
		if !p.lastSampleTimestamp.IsZero() && sample.Time.Sub(p.lastSampleTimestamp) >= p.FlushSampleLag {
			flush = true
		}
		p.lastSampleTimestamp = sample.Time
	}
	if flush {
		err = p.triggerFlush(oldHeader, false)
	}
	if p.FlushNoSampleTimeout > 0 {
		p.lastSample = time.Now()
		if err == nil {
			err = p.lastAutoFlushError
		}
		p.lastAutoFlushError = nil
	}
	p.samples = append(p.samples, sample)

	if p.ForwardImmediately {
		if err != nil {
			log.Errorln("Dropping error, because forward-immediately mode is enabled:", err)
		}
		return p.NoopProcessor.Sample(sample, header)
	} else {
		return
	}
}

func (p *BatchProcessor) Close() {
	defer p.NoopProcessor.Close()
	if !p.DontFlushOnClose {
		header := p.checker.LastHeader
		if header == nil {
			log.Warnln(p.String(), "received no samples")
		}
		if err := p.triggerFlush(header, true); err != nil {
			p.Error(err)
		}
	}
}

func (p *BatchProcessor) triggerFlush(header *Header, shutdown bool) error {
	p.flushTrigger.L.Lock()
	defer p.flushTrigger.L.Unlock()
	p.flushHeader = header
	p.flushTrigger.Broadcast()
	p.shutdown = shutdown
	for p.flushHeader != nil {
		p.flushTrigger.Wait() // Will be notified after flush is finished
	}
	res := p.flushError
	p.flushError = nil
	return res
}

func (p *BatchProcessor) loopFlush(wg *sync.WaitGroup) {
	defer wg.Done()
	for p.waitAndExecuteFlush() {
	}
}

func (p *BatchProcessor) waitAndExecuteFlush() bool {
	p.flushTrigger.L.Lock()
	defer p.flushTrigger.L.Unlock()
	for p.flushHeader == nil && !p.shutdown && !p.flushTimedOut() {
		if p.FlushNoSampleTimeout > 0 {
			p.flushTrigger.WaitTimeout(p.FlushNoSampleTimeout)
		} else {
			p.flushTrigger.Wait()
		}
	}
	if p.flushHeader == nil && !p.shutdown {
		// Automatic flush after timeout
		err := p.executeFlush(p.checker.LastHeader)
		if err != nil {
			log.Errorf("%v: Error during automatic flush (will be returned when next sample arrives): %v", p, err)
			p.lastAutoFlushError = fmt.Errorf("Error during previous auto-flush: %v", err)
		}
		p.lastSample = time.Now()
	} else {
		p.flushError = p.executeFlush(p.flushHeader)
		p.flushTrigger.Broadcast()
	}
	p.flushHeader = nil
	return !p.shutdown
}

func (p *BatchProcessor) flushTimedOut() bool {
	if p.FlushNoSampleTimeout <= 0 || p.lastSample.IsZero() {
		return false
	}
	return time.Now().Sub(p.lastSample) >= p.FlushNoSampleTimeout
}

func (p *BatchProcessor) executeFlush(header *Header) error {
	samples := p.samples
	if len(samples) == 0 || header == nil {
		return nil
	}
	p.samples = nil // Allow garbage collection
	if samples, header, err := p.executeSteps(samples, header); err != nil {
		return err
	} else if p.ForwardImmediately {
		log.Println("Dropping", len(samples), "batched samples (forward immediately mode is enabled)")
		return nil
	} else {
		if header == nil {
			return fmt.Errorf("Cannot flush %v samples because nil-header was returned by last batch processing step", len(samples))
		}
		if len(samples) > 0 {
			log.Println("Flushing", len(samples), "batched samples with", len(header.Fields), "metrics")
			for _, sample := range samples {
				if err := p.NoopProcessor.Sample(sample, header); err != nil {
					return fmt.Errorf("Error flushing batch: %v", err)
				}
			}
		}
		return nil
	}
}

func (p *BatchProcessor) executeSteps(samples []*Sample, header *Header) ([]*Sample, *Header, error) {
	if len(p.Steps) > 0 {
		log.Debugln("Executing", len(p.Steps), "batch processing step(s)")
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
	if len(p.FlushTags) > 0 {
		flushed = fmt.Sprintf(", flushed with tags %v", p.FlushTags)
	}
	if p.FlushNoSampleTimeout > 0 {
		flushed += fmt.Sprintf(", auto-flushed after %v", p.FlushNoSampleTimeout)
	}
	if p.FlushSampleLag > 0 {
		flushed += fmt.Sprintf(", flushed when sample timestamp difference over %v", p.FlushSampleLag)
	}
	return fmt.Sprintf("BatchProcessor (%v step%s%s)", len(p.Steps), extra, flushed)
}

func (p *BatchProcessor) MergeProcessor(other SampleProcessor) bool {
	if otherBatch, ok := other.(*BatchProcessor); ok && p.compatibleParameters(otherBatch) {
		p.Steps = append(p.Steps, otherBatch.Steps...)
		return true
	} else {
		return false
	}
}

func (p *BatchProcessor) compatibleParameters(other *BatchProcessor) bool {
	return reflect.DeepEqual(p.copyOnlyParameters(), other.copyOnlyParameters())
}

func (p *BatchProcessor) copyOnlyParameters() *BatchProcessor {
	return &BatchProcessor{
		// This list must remain synchronized with the actual exported fields
		ForwardImmediately:   p.ForwardImmediately,
		DontFlushOnClose:     p.DontFlushOnClose,
		FlushNoSampleTimeout: p.FlushNoSampleTimeout,
		FlushSampleLag:       p.FlushSampleLag,
		FlushAfterNumSamples: p.FlushAfterNumSamples,
		FlushAfterTime:       p.FlushAfterTime,
		FlushTags:            p.FlushTags,
	}
}

// ==================== Simple implementation ====================

type SimpleBatchProcessingStep struct {
	Description          string
	Process              func(header *Header, samples []*Sample) (*Header, []*Sample, error)
	OutputSampleSizeFunc func(sampleSize int) int
}

func (s *SimpleBatchProcessingStep) ProcessBatch(header *Header, samples []*Sample) (*Header, []*Sample, error) {
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
