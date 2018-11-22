package steps

import (
	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
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
