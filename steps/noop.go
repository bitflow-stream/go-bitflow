package steps

import (
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterNoop(b reg.ProcessorRegistry) {
	b.RegisterStep("noop",
		func(p *bitflow.SamplePipeline, _ map[string]string) error {
			p.Add(new(NoopProcessor))
			return nil
		},
		"Pass samples through without modification")
}

type NoopProcessor struct {
	bitflow.NoopProcessor
}

func (*NoopProcessor) String() string {
	return "noop"
}
