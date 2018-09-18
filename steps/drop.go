package steps

import (
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
)

func RegisterDrop(b reg.ProcessorRegistry) {
	b.RegisterAnalysis("drop",
		func(p *pipeline.SamplePipeline) {
			p.Add(&pipeline.SimpleProcessor{
				Description: "Drop all samples",
				Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
					return nil, nil, nil
				},
			})
		},
		"Drop all samples")
}
