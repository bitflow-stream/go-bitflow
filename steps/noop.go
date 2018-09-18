package steps

import (
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
)

func RegisterNoop(b reg.ProcessorRegistry) {
	b.RegisterAnalysis("noop",
		func(p *pipeline.SamplePipeline) {
			p.Add(new(NoopProcessor))
		},
		"Pass samples through without modification")
}

type NoopProcessor struct {
	bitflow.NoopProcessor
}

func (*NoopProcessor) String() string {
	return "noop"
}
