package bitflow

import (
	"fmt"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

type WindowHandler interface {
	ClearWindow() []*Sample
	MoveWindow() []*Sample
	Add(sample *Sample)
	FlushRequired() bool
}

type SampleWindowHandler struct {
	Size     int
	StepSize int

	// Current sample window
	Samples []*Sample
	// Helper to track window position in sample stream
	diff int
}

func (s *SampleWindowHandler) ClearWindow() []*Sample {
	s.diff = 0
	samples := s.Samples
	s.Samples = nil
	return samples
}

func (s *SampleWindowHandler) MoveWindow() []*Sample {
	s.diff = s.Size - s.StepSize
	if s.diff <= 0 {
		s.Samples = nil
	}
	samples := s.Samples
	// Remove first s.diff Samples
	return samples
}

func (s *SampleWindowHandler) Add(sample *Sample) {
	if s.diff >= 0 {
		s.Samples = append(s.Samples, sample)
	}
	s.diff++
}

func (s *SampleWindowHandler) FlushRequired() bool {
	return s.diff >= s.Size
}

type TimeWindowHandler struct {
	Size     time.Duration
	StepSize time.Duration

	// Current sample window
	samples []*Sample
	// Helper to track window position in sample stream
	diff time.Duration
}

func (t *TimeWindowHandler) ClearWindow() []*Sample {
	t.diff = time.Duration(0)
	samples :=  t.samples
	t.samples = nil
	return samples
}

func (t *TimeWindowHandler) MoveWindow() []*Sample {
	t.diff = t.Size - t.StepSize
	if t.diff <= 0 {
		t.samples = nil
	}

	samples := t.samples
	for i, sample := range t.samples {
		if i > 0 {
			difference := sample.Time.Sub(t.samples[0].Time)
			if difference >= t.StepSize {

				// Remove Samples until i (i is incl.)
			}
		}
	}
	return samples
}

func (t *TimeWindowHandler) Add(sample *Sample) {
	if t.diff >= 0 {
		t.samples = append(t.samples, sample)
	}
	if len(t.samples) > 1 {
		lastSample := t.samples[len(t.samples)-1]
		secondLastSample := t.samples[len(t.samples)-1]
		t.diff += lastSample.Time.Sub(secondLastSample.Time)
	}
}

func (t *TimeWindowHandler) FlushRequired() bool {
	return t.diff >= t.Size
}

type BatchProcessor struct {
	NoopProcessor
	checker HeaderChecker

	shutdown bool
	flushAll bool

	Steps []BatchProcessingStep

	// ForwardImmediately enables a mode where each received sample is immediately forwarded to the subsequent processing
	// step for further processing. The sample is also stored for batch processing, and all flush semantics are still active,
	// but the results of the batch processing will be dropped. Errors that occur during batch flushing are logged, but are
	// not reported to the caller of the Sample() method.
	ForwardImmediately bool

	// DontFlushOnClose disables the automatic final flushing when this processing step is closed. This is useful only in
	// certain special cases.
	DontFlushOnClose bool

	FlushNoSampleTimeout time.Duration // If > 0, flush when no new Samples are received for the given duration. The wall-time is used for this (not sample timestamps)
	FlushSampleLag       time.Duration // If > 0, flush when a sample is received with a timestamp jump bigger than this
	lastAutoFlushError   error
	lastSample           time.Time // Wall time when receiving last sample
	lastSampleTimestamp  time.Time // Timestamp of last sample

	Handler WindowHandler

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
	flushAll := false
	flush := p.Handler.FlushRequired()
	if p.checker.InitializedHeaderChanged(header) {
		flushAll = true
		flush = true
	}
	if len(p.FlushTags) > 0 {
		values := make([]string, len(p.FlushTags))
		for i, tag := range p.FlushTags {
			values[i] = sample.Tag(tag)
			if oldHeader != nil && len(p.lastFlushTags) > i && p.lastFlushTags[i] != values[i] {
				flush = true
				flushAll = true
			}
		}
		p.lastFlushTags = values
	}
	if p.FlushSampleLag > 0 {
		if !p.lastSampleTimestamp.IsZero() && sample.Time.Sub(p.lastSampleTimestamp) >= p.FlushSampleLag {
			flush = true
			flushAll = true
		}
		p.lastSampleTimestamp = sample.Time
	}
	if flush {
		err = p.triggerFlush(oldHeader, false, flushAll)
	}
	if p.FlushNoSampleTimeout > 0 {
		p.lastSample = time.Now()
		if err == nil {
			err = p.lastAutoFlushError
		}
		p.lastAutoFlushError = nil
	}
	p.Handler.Add(sample)

	if p.ForwardImmediately {
		if err != nil {
			log.Errorf("Dropping error, because forward-immediately mode is enabled: %v", err)
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
			log.Warnln(p.String(), "received no Samples")
		}
		if err := p.triggerFlush(header, true, true); err != nil {
			p.Error(err)
		}
	}
}

func (p *BatchProcessor) triggerFlush(header *Header, shutdown bool, flushAll bool) error {
	p.flushTrigger.L.Lock()
	p.flushAll = flushAll
	defer func() { p.flushAll = false }()
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
	var samples []*Sample
	if p.flushAll {
		samples = p.Handler.ClearWindow()
	} else {
		samples = p.Handler.MoveWindow()
	}
	if len(samples) == 0 || header == nil {
		return nil
	}

	if samples, header, err := p.executeSteps(samples, header); err != nil {
		return err
	} else if p.ForwardImmediately {
		log.Printf("Dropping %v batched Samples (forward immediately mode is enabled)", len(samples))
		return nil
	} else {
		if header == nil {
			return fmt.Errorf("Cannot flush %v Samples because nil-header was returned by last batch processing step", len(samples))
		}
		if len(samples) > 0 {
			log.Printf("Flushing %v Samples with %v metrics as result of the batch steps", len(samples), len(header.Fields))
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
				log.Warnln("Cannot execute remaining", len(p.Steps)-i, "batch step(s) because the batch with", len(header.Fields), "has no Samples")
				break
			} else {
				log.Println("Executing", step, "on", len(samples), "Samples with", len(header.Fields), "metrics")
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
	if (other.FlushNoSampleTimeout != 0 && other.FlushNoSampleTimeout != p.FlushNoSampleTimeout) ||
		(other.FlushSampleLag != 0 && other.FlushSampleLag != p.FlushSampleLag) {
		return false
	}
	if len(other.FlushTags) == 0 {
		return true
	}
	return golib.EqualStrings(p.FlushTags, other.FlushTags)
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
