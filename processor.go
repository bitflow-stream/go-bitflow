package pipeline

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
)

// MergeableProcessor is an extension of bitflow.SampleProcessor, that also allows
// merging two processor instances of the same time into one. Merging is only allowed
// when the result of the merge would has exactly the same functionality as using the
// two separate instances. This can be used as an optional optimization.
type MergeableProcessor interface {
	bitflow.SampleProcessor
	MergeProcessor(other bitflow.SampleProcessor) bool
}

// ==================== SimpleProcessor ====================

type SimpleProcessor struct {
	bitflow.NoopProcessor
	Description          string
	Process              func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error)
	OnClose              func()
	OutputSampleSizeFunc func(sampleSize int) int
}

func (p *SimpleProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if process := p.Process; process == nil {
		return fmt.Errorf("%s: Process function is not set", p)
	} else {
		sample, header, err := process(sample, header)
		if err == nil && sample != nil && header != nil {
			err = p.NoopProcessor.Sample(sample, header)
		}
		return err
	}
}

func (p *SimpleProcessor) Close() {
	if c := p.OnClose; c != nil {
		c()
	}
	p.NoopProcessor.Close()
}

func (p *SimpleProcessor) OutputSampleSize(sampleSize int) int {
	if f := p.OutputSampleSizeFunc; f != nil {
		return f(sampleSize)
	}
	return sampleSize
}

func (p *SimpleProcessor) String() string {
	if p.Description == "" {
		return "SimpleProcessor"
	} else {
		return p.Description
	}
}
