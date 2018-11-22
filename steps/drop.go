package steps

import (
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterDrop(b reg.ProcessorRegistry) {
	b.RegisterAnalysis("drop",
		func(p *bitflow.SamplePipeline) {
			p.Add(&bitflow.SimpleProcessor{
				Description: "Drop all samples",
				Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
					return nil, nil, nil
				},
			})
		},
		"Drop all samples")
}
