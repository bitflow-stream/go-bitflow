package steps

import (
	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

func RegisterNoop(b *query.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline) {
		p.Add(new(NoopProcessor))
	}
	b.RegisterAnalysis("noop", create, "Pass samples through without modification")
}

type NoopProcessor struct {
	bitflow.NoopProcessor
}

func (*NoopProcessor) String() string {
	return "noop"
}
