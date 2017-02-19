package pipeline

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
)

type SamplePipeline struct {
	bitflow.SamplePipeline
	lastProcessor bitflow.SampleProcessor
}

func (p *SamplePipeline) Add(step bitflow.SampleProcessor) *SamplePipeline {
	if p.lastProcessor != nil {
		if merger, ok := p.lastProcessor.(MergableProcessor); ok {
			if merger.MergeProcessor(step) {
				// Merge successful: drop the incoming step
				return p
			}
		}
	}
	p.lastProcessor = step
	p.SamplePipeline.Add(step)
	return p
}

func (p *SamplePipeline) Batch(steps ...BatchProcessingStep) *SamplePipeline {
	batch := new(BatchProcessor)
	for _, step := range steps {
		batch.Add(step)
	}
	return p.Add(batch)
}

func (p *SamplePipeline) String() string {
	return "Pipeline"
}

func (p *SamplePipeline) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, 0, len(p.Processors)+2)
	if p.Source != nil {
		res = append(res, p.Source)
	}
	for _, proc := range p.Processors {
		res = append(res, proc)
	}
	if p.Sink != nil {
		res = append(res, p.Sink)
	}
	return res
}

func (p *SamplePipeline) Format() []string {
	printer := IndentPrinter{
		OuterIndent:  "│ ",
		InnerIndent:  "├─",
		CornerIndent: "└─",
		FillerIndent: "  ",
	}
	return printer.PrintLines(p)
}
